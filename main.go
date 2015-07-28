package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/utils"

	"github.com/applidget/psdock/fsdriver"
	"github.com/applidget/psdock/logrotate"
	"github.com/applidget/psdock/notifier"
	"github.com/applidget/psdock/stream"
	"github.com/applidget/psdock/system"
)

const (
	containersRoot = "/var/run/psdock"
)

var (
	version     string // this variable is populated by the makefile
	standardEnv = &cli.StringSlice{
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		"TERM=xterm",
	}
)

func main() {
	app := cli.NewApp()
	app.Name = "psdock"
	app.Version = fmt.Sprintf("v%s", version)
	app.Author = "Applidget"
	app.Usage = "simple container engine"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "image, i", Usage: "container image"},
		cli.StringFlag{Name: "rootfs, r", Usage: "container rootfs"},
		cli.StringFlag{Name: "stdio", Usage: "standard input/output, if not specified, will use current stdin and stdout"},
		cli.StringFlag{Name: "stdout-prefix", Usage: "add a prefix to container output lines (format: <prefix>:<color>)"},
		cli.StringFlag{Name: "web-hook", Usage: "web hook to notify process status changes"},
		cli.StringFlag{Name: "bind-port", Usage: "port the process is expected to bind"},
		cli.StringFlag{Name: "user, u", Value: "root", Usage: "user inside container"},
		cli.StringFlag{Name: "cwd", Usage: "set the current working dir"},
		cli.StringFlag{Name: "hostname", Value: "psdock", Usage: "set the container hostname"},
		cli.StringSliceFlag{Name: "env, e", Value: standardEnv, Usage: "set environment variables for the process"},
		cli.StringSliceFlag{Name: "bind-mount", Value: &cli.StringSlice{}, Usage: "set bind mounts"},
		cli.IntFlag{Name: "log-rotate", Usage: "rotate stdout output (if stdio is a proper file)"},
		cli.IntFlag{Name: "kill-timeout", Value: -1, Usage: "kill the process after timeout after receiving a SIGINT or SIGTERM"},
	}
	app.Commands = []cli.Command{
		cli.Command{
			Name:   "init",
			Usage:  "container init, should never be invoked manually",
			Action: initAction,
		},
	}
	app.Action = func(c *cli.Context) {
		exit, err := start(c)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(exit)
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func initAction(c *cli.Context) {
	runtime.GOMAXPROCS(1)
	runtime.LockOSThread()

	factory, err := libcontainer.New("")
	if err != nil {
		log.Fatal(err)
	}
	if err := factory.StartInitialization(); err != nil {
		log.Fatal(err)
	}
	panic("This line should never been executed")
}

func start(c *cli.Context) (int, error) {
	// setup rootfs
	image := c.String("image")
	if image == "" {
		return 1, fmt.Errorf("no image specified")
	}
	image = path.Clean(image)

	rootfs, _ := filepath.Abs(c.String("rootfs"))
	if rootfs == "" {
		return 1, fmt.Errorf("no rootfs specified")
	}
	rootfs = path.Clean(rootfs)

	driver, err := fsdriver.New(image, rootfs)
	if err != nil {
		return 1, err
	}

	if err := driver.SetupRootfs(); err != nil {
		return 1, err
	}
	defer driver.CleanupRootfs()

	// create container factory
	bin, err := exec.LookPath("psdock")
	if err != nil {
		//psdock not in the path
		bin, _ = filepath.Abs(os.Args[0])
	}
	factory, err := libcontainer.New(containersRoot, libcontainer.InitArgs(bin, "init"), libcontainer.Cgroupfs)
	if err != nil {
		return 1, err
	}

	// create container
	cuid, _ := utils.GenerateRandomName("psdock_", 7)
	config, err := loadConfig(cuid, rootfs, c.String("hostname"), c.StringSlice("bind-mount"))
	if err != nil {
		return 1, err
	}

	container, err := factory.Create(cuid, config)
	if err != nil {
		return 1, err
	}
	defer container.Destroy()

	//write PID of launching process, it will be next to the state.json file
	if err := ioutil.WriteFile(filepath.Join(containersRoot, cuid, "pid"), []byte(fmt.Sprintf("%d", os.Getpid())), 0600); err != nil {
		return 1, err
	}

	// prepare process
	process := &libcontainer.Process{
		Args: c.Args(),
		Env:  c.StringSlice("env"),
		User: c.String("user"),
		Cwd:  c.String("cwd"),
	}

	// prepare stdio stream
	pref, prefColor := parsePrefixArg(c.String("stdout-prefix"))
	s, err := stream.NewStream(c.String("stdio"), pref, prefColor)
	if err != nil {
		return 1, err
	}
	defer s.Close()

	var tty *tty
	if !s.Interactive() {
		//no tty
		process.Stdin = nil
		process.Stdout = s
		process.Stderr = s
	} else {
		rootuid, err := config.HostUID()
		if err != nil {
			return 1, err
		}

		tty, err = newTty(process, rootuid)
		if err != nil {
			return 1, err
		}

		if err := tty.attach(s); err != nil {
			return 1, err
		}
		tty.resize()

		defer tty.Close()
	}

	//setup log rotation if wanted
	if c.Int("log-rotate") > 0 && s.URL.Scheme == "file" {
		r := logrotate.NewRotator(s.URL.Host + s.URL.Path)
		r.RotationDelay = time.Duration(c.Int("log-rotate")) * time.Hour
		go r.StartWatching()
		defer r.StopWatching()
	}

	// forward received signals to container process
	signalHandler := &signalHandler{process: process, tty: tty, killTimeout: c.Int("kill-timeout")}
	go signalHandler.startCatching()

	if s.Interactive() {
		go func() {
			<-s.CloseCh
			//if interactive and stream closed, send a sigterm to the process
			signalHandler.handleInterupt(syscall.SIGTERM)
		}()
	}

	// start container process
	statusChanged(c, notifier.StatusStarting)
	defer statusChanged(c, notifier.StatusCrashed)

	// start the container
	if err := container.Start(process); err != nil {
		return 1, err
	}

	if c.String("bind-port") == "" {
		statusChanged(c, notifier.StatusRunning)
	} else {
		go func() {
			port := c.String("bind-port")
			for {
				//
				pids, err := container.Processes()
				if err != nil {
					log.Errorf("failed to get back container processes: %v", err)
					// if this arise, we just do not change process status
					return
				}

				bound, err := system.IsPortBound(port, pids)
				if err != nil || !bound {
					if err != nil {
						log.Errorf("failed to check if port %s is bound: %v", port, err)
					}
					//will retry
					time.Sleep(200 * time.Millisecond)
				} else {
					break
				}
			}

			statusChanged(c, notifier.StatusRunning)
		}()
	}

	// container exited
	status, err := process.Wait()

	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if ok {
			status = exitError.ProcessState
		} else {
			return 1, err
		}
	}

	exit := utils.ExitStatus(status.Sys().(syscall.WaitStatus))
	if signalHandler.forceKilled && exit == 137 { //128 + 9 (kill) indicates a kill exit status
		//sigterm sent to process but was converted to a sigkill so assume no errors
		return 0, nil
	}

	return exit, nil
}

// call webhook if needed
func statusChanged(c *cli.Context, status notifier.PsStatus) {
	wh := c.String("web-hook")
	if wh == "" {
		return
	}
	notifier.WebHook = wh

	if err := notifier.NotifyHook(status); err != nil {
		log.Error("failed to notify web-hook %s: %v", wh, err)
	}
}
