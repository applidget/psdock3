package integration

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
)

// struct to play with the psdock binary
type binary struct {
	name   string
	ps     *os.Process
	stdout []byte
	stderr []byte
}

func newBinary() *binary {
	return &binary{name: "../psdock"}
}

func (b *binary) start(args ...string) error {
	cmd := exec.Command(b.name, args...)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return err
	}
	b.ps = cmd.Process
	b.stdout, _ = ioutil.ReadAll(stdout)
	b.stderr, _ = ioutil.ReadAll(stderr)
	return cmd.Wait()
}

func (b *binary) stop() error {
	return b.ps.Signal(syscall.SIGTERM)
}

func (b *binary) debugInfo() string {
	return fmt.Sprintf("stdout: %s\nstderr: %s", string(b.stdout), string(b.stderr))
}
