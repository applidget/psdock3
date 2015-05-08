package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/libcontainer"
	"github.com/docker/libcontainer/utils"

	"github.com/robinmonjo/psdock/fsdriver"
	_ "github.com/robinmonjo/psdock/logrotate"
	"github.com/robinmonjo/psdock/notifier"
	"github.com/robinmonjo/psdock/portwatcher"
	"github.com/robinmonjo/psdock/stream"
)

const (
	version             = "0.1"
	libcontainerVersion = "b6cf7a6c8520fd21e75f8b3becec6dc355d844b0"
)

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

var (
	standardEnv = []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin", "HOSTNAME=psdock", "TERM=xterm"}
)

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

	config := loadConfig(rootfs)
	container, err := factory.Create("psdock", config)
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

		defer tty.Close()
	}

	//log.Infof("Stream is interactive: %v, is terminal: %v", s.Interactive(), s.Terminal())

	// forward received signals to container process
	signalHandler := &signalHandler{container: container, process: process, tty: tty}
	go signalHandler.startCatching()

	if s.Interactive() {
		go func() {
			<-s.CloseCh
			//if interactive and stream closed, send a sigterm to the process
			signalHandler.handleInterupt(syscall.SIGTERM)
		}()
	}

	// start container process
	statusChanged := func(status notifier.PsStatus) {
		wh := c.GlobalString("webhook")
		if wh == "" {
			return
		}
		notifier.WebHook = wh

		if err := notifier.NotifyHook(status); err != nil {
			log.Error("failed to notify web hook %s: %v", wh, err)
		}
	}

	statusChanged(notifier.StatusStarting)

	if c.GlobalString("bindport") == "" {
		statusChanged(notifier.StatusRunning)
	} else {
		go func() {
			//wait until we have a pid and until the port is bound
			pid, err := initProcessPid(container)
			if err != nil {
				log.Errorf("unable to get back container init process pid: %v", err)
				return
			}
			if _, err := portwatcher.Watch(pid, c.GlobalString("bindport")); err != nil {
				log.Errorf("failed to watch port: %v", err)
				return
			}
			//at this point the process has bound the port
			statusChanged(notifier.StatusRunning)
		}()
	}

	if err := container.Start(process); err != nil {
		return 1, err
	}

	status, err := process.Wait()
	statusChanged(notifier.StatusCrashed)

	if err != nil {
		return 1, err
	}
	return utils.ExitStatus(status.Sys().(syscall.WaitStatus)), nil
}
