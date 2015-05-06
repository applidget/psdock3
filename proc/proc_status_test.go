package proc

import (
	"fmt"
	"os/exec"
	"sync"
	"testing"
)

func Test_decodeMask(t *testing.T) {
	masks := []string{"fffffffe7ffbfeff", "00000000280b2603", "0000000000000000"}
	expected := [][]int{
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 10, 11, 12, 13, 14, 15, 16, 17, 18, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
		[]int{1, 2, 10, 11, 14, 17, 18, 20, 28, 30},
		[]int{},
	}

	for i, mask := range masks {
		signals, err := decodeMask(mask)
		if err != nil {
			t.Fatal(err)
		}
		exp := expected[i]
		if len(exp) != len(signals) {
			t.Fatalf("expected %d signals got %d for mask %s", len(exp), len(signals), mask)
		}

		for j, sig := range signals {
			if int(sig) != exp[j] {
				t.Fatalf("expected sig %d, got sig number %d", sig, exp[j])
			}
		}
	}
}

func Test_procInfo(t *testing.T) {

	var wg sync.WaitGroup
	wg.Add(1)

	cmd := exec.Command("tail", "-f")

	go func() {
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		wg.Done()
		if err := cmd.Wait(); err != nil {
			t.Fatal(err)
		}
	}()

	wg.Wait()

	pid := fmt.Sprintf("%d", cmd.Process.Pid)
	defer cmd.Process.Kill()

	ps, err := NewProcStatus(pid)
	if err != nil {
		t.Fatal(err)
	}

	if ps.Pid != pid {
		t.Fatalf("expected pid %s got %s", pid, ps.Pid)
	}
}