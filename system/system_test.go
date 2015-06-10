package system

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

func Test_portBinder(t *testing.T) {
	fmt.Printf("port binder ... ")
	port, err := freePort()
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("nc", "-l", port)
	err = cmd.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	time.Sleep(100 * time.Millisecond) //just make sure cmd has time to bind the port

	binder, err := portBinder(port)
	if err != nil {
		t.Fatal(err)
	}

	if binder != cmd.Process.Pid {
		t.Fatalf("wrong port binder, expected %d got %d", cmd.Process.Pid, binder)
	}
	fmt.Println("done")
}

func Test_isPortBoundFalse(t *testing.T) {
	fmt.Printf("is port bound faillure ... ")
	pids := []int{2344, 2445, 1}
	bound, err := IsPortBound("9999", pids) // none of these pid should have bound the 9999 port
	if err != nil {
		t.Fatal(err)
	}
	if bound {
		t.Fatalf("port 9999 must not be reported as bound by one of these processes %v", pids)
	}
	fmt.Println("done")
}

func Test_isPortBoundTrue(t *testing.T) {
	fmt.Printf("is port bound success ... ")
	cmd := exec.Command("nc", "-l", "9999")
	err := cmd.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	pid := fmt.Sprintf("%d", cmd.Process.Pid)

	p, _ := strconv.Atoi(pid)

	pids := []int{2344, 2445, 1, p, 890}
	bound, err := IsPortBound("9999", pids)
	if err != nil {
		t.Fatal(err)
	}
	if !bound {
		t.Fatalf("port 9999 should be reported has bound by %d", p)
	}
	fmt.Println("done")
}

//helper

//return a free port
func freePort() (string, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", err
	}
	defer l.Close()
	return strings.TrimPrefix(l.Addr().String(), "[::]:"), nil
}
