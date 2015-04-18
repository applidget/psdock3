package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/libcontainer"
	"github.com/docker/libcontainer/utils"
)

const (
	//versions
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
		cli.StringFlag{Name: "uid", Usage: "container unique id"},
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

	//mount image with overlayfs and use it as rootfs + defer umount
	rootfs := c.GlobalString("image") //for now running in image directly

	bin, _ := filepath.Abs(os.Args[0])
	factory, err := libcontainer.New(rootfs, libcontainer.InitArgs(bin, "init"))
	if err != nil {
		return 1, err
	}

	uid := c.GlobalString("uid")
	container, err := factory.Create(uid, loadConfig(uid, rootfs))
	if err != nil {
		return 1, err
	}
	defer container.Destroy()
	process := &libcontainer.Process{
		Args:   []string{"bash"},
		Env:    []string{"PATH=/usr/local/bin:/bin"},
		User:   "root",
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	//launc every co-process
	go handleSignals(process)

	if err := container.Start(process); err != nil {
		return 1, err
	}

	// wait for the process to finish.
	status, err := process.Wait()
	if err != nil {
		return 1, err
	}

	return utils.ExitStatus(status.Sys().(syscall.WaitStatus)), nil
}

//co processes
func handleSignals(container *libcontainer.Process) {
	sigc := make(chan os.Signal, 10)
	signal.Notify(sigc)
	for sig := range sigc {
		container.Signal(sig)
	}
}

func listenSocketBinding(port string) {

}
