package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/docker/libcontainer"
	"github.com/robinmonjo/psdock/stream"
)

// prefix args have the following format: --prefix some-prefix[:blue]
func parsePrefixArg(prefix string) (string, stream.Color) {
	comps := strings.Split(prefix, ":")
	if len(comps) == 1 {
		return comps[0], stream.NoColor
	}
	return comps[0], stream.MapColor(comps[len(comps)-1])
}

// returns pid of the init process of a container, from the host point of view
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
		return initProcessPid(container) //wait until the state knows the PID
	}
}
