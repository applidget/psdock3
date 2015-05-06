package main

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/libcontainer"
	"github.com/docker/libcontainer/utils"

	"github.com/robinmonjo/psdock/coprocs"

	"github.com/robinmonjo/psdock/fsdriver"
	"github.com/robinmonjo/psdock/notifier"
	"github.com/robinmonjo/psdock/proc"
	"github.com/robinmonjo/psdock/stream"
)

const (
	version             = "0.1"
	libcontainerVersion = "b6cf7a6c8520fd21e75f8b3becec6dc355d844b0"
)

var standardEnv = []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin", "HOSTNAME=psdock", "TERM=xterm"}

func init() {
	env := os.Getenv("GO_ENV")
	if env == "testing" {
		log.SetLevel(log.DebugLevel)
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "psdock"
	app.Version = fmt.Sprintf("v%s (libcontainer %s)", version, libcontainerVersion)
	app.Author = "Robin Monjo"
	app.Email = "robinmonjo@gmail.com"
	app.Usage = "simple container engine specialized in PaaS"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "image, i", Usage: "container image"},
		cli.StringFlag{Name: "rootfs, r", Usage: "container image"},
		cli.StringFlag{Name: "stdio", Usage: "standard input/output connection, if not specified, will use os stdin and stdout"},
		cli.StringFlag{Name: "prefix", Usage: "add a prefix to container output lines (format: <prefix>:<color>)"},
		cli.StringFlag{Name: "webhook, wh", Usage: "web hook to notify process status"},
		cli.StringFlag{Name: "binport, bp", Usage: "port the process is expected to bind"},
		cli.StringFlag{Name: "user, u", Value: "root", Usage: "user inside container"},
		cli.StringFlag{Name: "cwd", Usage: "set the current working dir"},
		cli.IntFlag{Name: "kill-timeout", Value: 10, Usage: "time to wait for process gracefull stop before killing it"},
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
	// Mount usable rootfs (image are immutable)
	image := c.GlobalString("image")
	if image == "" {
		return 1, fmt.Errorf("no image specified")
	}
	image = path.Clean(image)

	rootfs := c.GlobalString("rootfs")
	if rootfs == "" {
		return 1, fmt.Errorf("no rootfs specified")
	}
	rootfs = path.Clean(rootfs)

	overlay, err := fsdriver.NewOverlay(image, rootfs)
	if err != nil {
		return 1, err
	}

	if err := overlay.SetupRootfs(); err != nil {
		return 1, err
	}
	defer overlay.CleanupRootfs()

	// package the app into the rootfs (bindmount ? cp ? possibility to have multi packer)

	// Configure the container and its process
	bin, _ := filepath.Abs(os.Args[0])
	factory, err := libcontainer.New(rootfs, libcontainer.InitArgs(bin, "init"))
	if err != nil {
		return 1, err
	}

	uid, err := utils.GenerateRandomName("psdock_", 7)
	if err != nil {
		return 1, err
	}
	config := loadConfig(uid, rootfs)
	container, err := factory.Create(uid, config)
	if err != nil {
		return 1, err
	}
	defer container.Destroy()

	process := &libcontainer.Process{
		Args: c.Args(),
		Env:  append(standardEnv, []string{}...),
		User: c.GlobalString("user"),
		Cwd:  c.GlobalString("cwd"),
	}

	pref, prefColor := parsePrefixArg(c.GlobalString("prefix"))
	s, err := stream.NewStream(c.GlobalString("stdio"), pref, prefColor)
	if err != nil {
		return 1, err
	}
	defer s.Close()

	var tty *tty
	if !s.Interactive() {
		//no tty
		process.Stdin = os.Stdin
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

		//at this point os.Stdout might not be usable anymore so logs of logrus wont work
		// I would bufferise them in a file and before exiting, output them (after tty.Close())
		defer tty.Close()
	}

	// forward received signals to container process
	go handleSignals(process, container, tty)

	// start container process
	psStatusChanged(c, notifier.StatusStarting)

	go monitorContainerStartup(container, c)

	if err := container.Start(process); err != nil {
		return 1, err
	}

	status, err := process.Wait()
	psStatusChanged(c, notifier.StatusCrashed)

	if err != nil {
		return 1, err
	}
	return utils.ExitStatus(status.Sys().(syscall.WaitStatus)), nil
}

func handleSignals(process *libcontainer.Process, container libcontainer.Container, tty *tty) {
	sigc := make(chan os.Signal, 10)
	signal.Notify(sigc)
	for sig := range sigc {
		log.Infof("signal received: %v", sig)
		if sig == syscall.SIGWINCH && tty != nil {
			tty.resize()
		} else if sig == syscall.SIGINT || sig == syscall.SIGTERM {
			// init process will have PID 1 in the namespace and by default PID 1 ignore all signals (https://github.com/docker/docker/issues/7846)
			// expect sigkill of course. Solution: inspect signal status (/proc/PID/signal), if it doesn't handle any signals, kill it otherwise
			// just forward the signal

			// if sigint or sigterm, check if the signal can caught them, if yes, send it otherwise kill the process (SIGSTOP and SIGKILL can't be caught)
			log.Info("will handle the shit")
			pid, err := initProcessPid(container)
			if err != nil {
				log.Errorf("error getting back PID: %v", err)
			} else {
				log.Info("PID is: %d", pid)
			}

			ps, err := proc.NewProcStatus(pid)
			if err != nil {
				log.Error(err)
				process.Signal(sig)
				return
			}

			if ps.SignalCaught(sig.(syscall.Signal)) {
				log.Infof("signal is caught so let the thing handle it: %v", sig)
				process.Signal(sig)
			} else {
				log.Infof("signal is not caught, killing: %v", sig)
				process.Signal(syscall.SIGKILL)
			}

		} else {
			//forward the signal
			process.Signal(sig)
		}
	}
}

//helpers
func parsePrefixArg(prefix string) (string, stream.Color) {
	comps := strings.Split(prefix, ":")
	if len(comps) == 1 {
		return comps[0], stream.NoColor
	}
	return comps[0], stream.MapColor(comps[len(comps)-1])
}

func monitorContainerStartup(container libcontainer.Container, c *cli.Context) {
	if c.GlobalString("bindport") == "" {
		//container process not expecting to bind a port
		psStatusChanged(c, notifier.StatusRunning)
		return
	}

	pid, err := initProcessPid(container)
	if err != nil {
		log.Errorf("unable to get back container init process pid: %v", err)
		return
	}

	if _, err := coprocs.Watch(pid, c.GlobalString("bindport")); err != nil {
		log.Errorf("failed to watch port: %v", err)
		return
	}
	//at this point the process has bound the port
	psStatusChanged(c, notifier.StatusRunning)
}

func psStatusChanged(c *cli.Context, status notifier.PsStatus) {
	wh := c.GlobalString("webhook")
	if wh == "" {
		return
	}
	notifier.WebHook = wh

	if err := notifier.NotifyHook(status); err != nil {
		log.Error("failed to notify web hook %s: %v", wh, err)
	}
}

func initProcessPid(container libcontainer.Container) (string, error) {
	retryDelay := 100 * time.Millisecond

	state, err := container.State()
	if err != nil {
		if err.(libcontainer.Error).Code() == libcontainer.ContainerNotExists {
			time.Sleep(retryDelay)
			return initProcessPid(container) //wait until the state exists
		} else {
			return "", err
		}
	}

	//state exists
	if state.InitProcessPid != 0 {
		return fmt.Sprintf("%d", state.InitProcessPid), nil
	} else {
		time.Sleep(retryDelay)
		return initProcessPid(container) //wait until the state exists
	}
}
