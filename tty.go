package main

//totally inspired by https://github.com/opencontainers/runc/libcontainer/blob/master/nsinit/tty.go

import (
	"io"
	"os"

	"github.com/applidget/psdock/stream"
	"github.com/docker/docker/pkg/term"
	"github.com/opencontainers/runc/libcontainer"
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
	// copy console output to stream's stdout
	go func() {
		io.Copy(s, t.console)
		s.Close()
	}()

	if s.Input == nil {
		return nil
	}

	// copy stream's stdin into the console
	go func() {
		io.Copy(t.console, s)
		s.Close()
	}()

	if s.Input == os.Stdin {
		// the current terminal shall pass everything to the console, make it ignores ctrl+C etc ...
		// this is done by making the terminal raw. The state is saved to reset user's terminal settings
		// when psdock exits
		state, err := term.SetRawTerminal(os.Stdin.Fd())
		if err != nil {
			return err
		}
		t.state = state
	} else {
		// s.Input is a socket (tcp, tls ...). Obvioulsy, we can't set the remote user's terminal in raw mode, however we can at least
		// disable echo on the console
		state, err := term.SaveState(t.console.Fd())
		if err != nil {
			return err
		}
		if err := term.DisableEcho(t.console.Fd(), state); err != nil {
			return err
		}
		t.state = state
	}

	return nil
}

func (t *tty) resize() error {
	ws, err := term.GetWinsize(os.Stdin.Fd())
	if err != nil {
		return err
	}
	return term.SetWinsize(t.console.Fd(), ws)
}
