package main

//totally inspired by https://github.com/docker/libcontainer/blob/master/nsinit/tty.go

import (
	"io"
	"os"

	"github.com/docker/docker/pkg/term"
	"github.com/docker/libcontainer"
	"github.com/robinmonjo/psdock/stream"
)

type tty struct {
	console libcontainer.Console
	state   *term.State
}

func newTty(p *libcontainer.Process, rootuid int) (*tty, error) {
	console, err := p.NewConsole(rootuid)
	if err != nil {
		return nil, err
	}
	return &tty{console: console}, nil
}

func (t *tty) Close() error {
	if t.console != nil {
		t.console.Close()
	}
	if t.state != nil {
		term.RestoreTerminal(os.Stdin.Fd(), t.state)
	}
	return nil
}

func (t *tty) attach(s *stream.Stream) error {
	if s.Input != nil { //stdin might be nil if stdio is a file
		go io.Copy(t.console, s)
	}
	go io.Copy(s, t.console)
	go io.Copy(s, t.console)

	if s.Input == nil {
		return nil
	}

	state, err := term.SetRawTerminal(t.console.Fd())
	if err != nil {
		return err
	}
	t.state = state
	return nil
}

func (t *tty) resize() error {
	ws, err := term.GetWinsize(os.Stdin.Fd())
	if err != nil {
		return err
	}
	return term.SetWinsize(t.console.Fd(), ws)
}
