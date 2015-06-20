package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/applidget/psdock/system"
	"github.com/docker/libcontainer"
)

type signalHandler struct {
	container   libcontainer.Container
	process     *libcontainer.Process
	tty         *tty
	forceKilled bool
}

func (h *signalHandler) startCatching() {
	sigc := make(chan os.Signal, 10)
	signal.Notify(sigc)

	for sig := range sigc {
		switch sig {
		case syscall.SIGWINCH:
			h.handleSigwinch()
		case syscall.SIGTERM:
			fallthrough
		case syscall.SIGINT:
			h.handleInterupt(sig)
		default:
			h.handleDefault(sig)
		}
	}
}

// terminal resize signal
func (h *signalHandler) handleSigwinch() {
	if h.tty != nil {
		h.tty.resize()
	}
}

// handle sigterm and sigint
func (h *signalHandler) handleInterupt(sig os.Signal) error {
	// init process will have PID 1 in the namespace and by default PID 1 ignore all signals (https://github.com/docker/docker/issues/7846)
	// expect sigkill of course. Solution: inspect signal status (/proc/PID/signal), if it doesn't handle any signals, kill it otherwise
	// just forward the signal

	// if sigint or sigterm, check if the signal can caught them, if yes, send it otherwise kill the process (SIGSTOP and SIGKILL can't be caught)
	pid, err := initProcessPid(h.container)
	if err != nil {
		//couldn't get pid, fallabck (probably the process died, already, anyway falling back to default)
		return h.handleDefault(sig)
	}

	ps, err := system.NewProcStatus(pid)
	if err != nil {
		log.Error(err)
		//failed to get process status, fallback
		return h.handleDefault(sig)
	}

	if ps.SignalCaught(sig.(syscall.Signal)) {
		log.Infof("signal is caught so let the thing handle it: %v", sig)
		return h.handleDefault(sig)
	}

	log.Infof("signal is not caught, killing: %v", sig)
	h.forceKilled = true
	return h.process.Signal(syscall.SIGKILL)
}

func (h *signalHandler) handleDefault(sig os.Signal) error {
	return h.process.Signal(sig)
}
