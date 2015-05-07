package portwatcher

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"
)

func Test_portBinder(t *testing.T) {
	fmt.Printf("port binder ... ")
	cmd := exec.Command("nc", "-l", "9999")
	err := cmd.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	pid := fmt.Sprintf("%d", cmd.Process.Pid)
	binder, err := portBinder("9999")

	if binder != pid {
		t.Fatalf("wrong port binder, expected %s got %s", pid, binder)
	}
	fmt.Println("done")
}

func Test_watch(t *testing.T) {
	fmt.Printf("port watch ... ")
	cmd := exec.Command("bash", "./assets/spawn.sh")
	cmd.Env = append(cmd.Env, "PORT=9090")
	err := cmd.Start()
	if err != nil {
		t.Fatal(err)
	}

	pid := fmt.Sprintf("%d", cmd.Process.Pid)

	binder, err := Watch(pid, "9090")
	if err != nil {
		t.Fatal(err)
	}

	if binder == pid {
		//shoudn't be the case
		t.Fatalf("port binder pid %s is equal to launcher pid %s", binder, pid)
	}

	//cleanup
	p, _ := strconv.Atoi(binder)
	procs, err := os.FindProcess(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := procs.Kill(); err != nil {
		t.Fatal(err)
	}
	fmt.Println("done")
}
