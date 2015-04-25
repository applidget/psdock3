package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/libcontainer"
	"github.com/docker/libcontainer/utils"

	_ "github.com/robinmonjo/psdock/coprocs"

	"github.com/robinmonjo/psdock/fsdriver"
	"github.com/robinmonjo/psdock/stream"
)

const (
	version             = "0.1"
	libcontainerVersion = "b6cf7a6c8520fd21e75f8b3becec6dc355d844b0"
)

var standardEnv = []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin", "HOSTNAME=nsinit", "TERM=xterm"}

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

	// Add google dns for networking
	if err := ioutil.WriteFile(filepath.Join(rootfs, "etc", "resolv.conf"), []byte("nameserver 8.8.8.8\nnameserver 4.4.4.4"), 0600); err != nil {
		return 1, err
	}

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
		//not setting stdin, stdout and stderr, we use a tty by default
	}

	// setup tty and and std{in, out, err} redirection + prefix
	rootuid, err := config.HostUID()
	if err != nil {
		return 1, err
	}
	tty, err := newTty(process, rootuid)
	if err != nil {
		return 1, err
	}

	pref, prefColor := parsePrefixArg(c.GlobalString("prefix"))
	s, err := stream.NewStream(c.GlobalString("stdio"), pref, prefColor)
	if err != nil {
		return 1, err
	}
	defer s.Close()

	if err := tty.attach(s); err != nil {
		return 1, err
	}
	defer tty.Close()

	// forward received signals to container process
	go handleSignals(process, tty)

	// launch co processes
	if c.GlobalString("bindport") != "" {
		//must be called inside the namespace ...
	}

	// start container process
	psStatusChanged(c, statusStarting)
	if err := container.Start(process); err != nil {
		return 1, err
	}

	status, err := process.Wait()
	if err != nil {
		return 1, err
	}

	// container's done
	psStatusChanged(c, statusCrashed)
	return utils.ExitStatus(status.Sys().(syscall.WaitStatus)), nil
}

func handleSignals(container *libcontainer.Process, tty *tty) {
	sigc := make(chan os.Signal, 10)
	signal.Notify(sigc)
	tty.resize()
	for sig := range sigc {
		switch sig {
		case syscall.SIGWINCH:
			tty.resize()
		default:
			container.Signal(sig)
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

func psStatusChanged(c *cli.Context, status psStatus) {
	wh := c.GlobalString("webhook")
	if wh == "" {
		return
	}

	if err := notifyHook(status); err != nil {
		log.Error("failed to notify web hook %s: %v", wh, err)
	}
}
