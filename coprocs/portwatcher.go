package coprocs

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

var (
	RetryDelay time.Duration = 200 //Milliseconds
)

func Watch(pid, port string) (string, error) {
	//wait until port is binded by pid or one of its children
	binderPid, err := portBinder(port)
	if err != nil {
		return "", err
	}

	if binderPid == "" {
		time.Sleep(RetryDelay * time.Millisecond)
		return Watch(pid, port)
	}

	pids, err := children(pid)
	if err != nil {
		return "", err
	}

	pids = append(pids, pid)
	for _, p := range pids {
		if p == binderPid {
			return binderPid, nil
		}
	}

	time.Sleep(RetryDelay * time.Millisecond)
	return Watch(pid, port)
}

// children return children and grand children of the given pid.
func children(pid string) ([]string, error) {
	out, err := exec.Command("pgrep", "-P", pid).Output()
	if err != nil {
		if exitStatus(err) == 1 {
			//no children
			return []string{}, nil
		}
		return nil, err
	}

	cpids := strings.Split(string(out), "\n")

	if cpids[len(cpids)-1] == "" {
		cpids = cpids[:len(cpids)-1] //remove the last empty line
	}

	for _, cpid := range cpids {
		ccpids, err := children(cpid)
		if err != nil {
			return nil, err
		}
		cpids = append(cpids, ccpids...)
	}

	return cpids, nil
}

func portBinder(port string) (string, error) {
	out, err := exec.Command("lsof", "-ti", fmt.Sprintf(":%s", port)).Output()
	if err != nil {
		if exitStatus(err) == 1 {
			//not binded
			return "", nil
		}
		return "", err
	}
	return strings.TrimSuffix(string(out), "\n"), nil
}

func exitStatus(err error) int {
	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}
	return 555
}
