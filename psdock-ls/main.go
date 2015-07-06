package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/opencontainers/runc/libcontainer"
)

const (
	containersRoot = "/var/run/psdock"
	stateFile      = "state.json"
)

type containerState struct {
	libcontainer.State
	launcherProcessPid string
}

var states []*containerState

func main() {
	listStates()

	if len(states) == 0 {
		fmt.Println("No psdock container running")
		return
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 10, 3, '\t', 0)

	fmt.Fprintln(tw, "#\tCONTAINER_ID\tPID\tINIT_PID\tROOTFS\tCOMMAND")
	for i, state := range states {
		runningCmd, err := cmdForPid(state.InitProcessPid)
		if err != nil {
			runningCmd = "unknown"
		}
		line := fmt.Sprintf("%d\t%s\t%s\t%d\t%s\t%s", i, state.ID, state.launcherProcessPid, state.InitProcessPid, state.Config.Rootfs, runningCmd)
		fmt.Fprintln(tw, line)
		tw.Flush()
	}
}

func visit(path string, fi os.FileInfo, err error) error {
	if path == containersRoot {
		return nil
	}

	if filepath.Base(path) != stateFile {
		return nil
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var state libcontainer.State
	if err := json.NewDecoder(f).Decode(&state); err != nil {
		return err
	}

	pidFile := filepath.Join(filepath.Dir(path), "pid")
	launcherPid := "-1"
	b, err := ioutil.ReadFile(pidFile)
	if err == nil {
		launcherPid = string(b)
	}

	states = append(states, &containerState{state, launcherPid})
	return nil
}

func listStates() {
	err := filepath.Walk(containersRoot, visit)
	if err != nil {
		panic(err)
	}
}

func cmdForPid(pid int) (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		return "", err
	}
	return string(b), nil
}
