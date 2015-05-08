package fsdriver

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// Driver create a usable rootfs from an imutable image directory
type Driver interface {
	SetupRootfs() error
	CleanupRootfs() error
}

type Overlay struct {
	lowerDir string
	upperDir string
	workDir  string
}

func NewOverlay(image, dest string) (*Overlay, error) {
	if err := supportsOverlay(); err != nil {
		return nil, err
	}
	workDir := fmt.Sprintf("%s_work", dest)
	return &Overlay{lowerDir: image, upperDir: dest, workDir: workDir}, nil
}

func (o *Overlay) SetupRootfs() error {
	//mount image in readonly into dest
	if err := os.MkdirAll(o.upperDir, 0700); err != nil {
		return err
	}
	if err := os.MkdirAll(o.workDir, 0700); err != nil {
		return err
	}
	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", o.lowerDir, o.upperDir, o.workDir)
	return syscall.Mount("overlay", o.upperDir, "overlay", 0, opts)
}

func (o *Overlay) CleanupRootfs() error {
	if err := syscall.Unmount(o.upperDir, 0); err != nil {
		return err
	}
	if err := os.RemoveAll(o.upperDir); err != nil {
		return err
	}
	return os.RemoveAll(o.workDir)
}

func supportsOverlay() error {
	exec.Command("modprobe", "overlay").Run()

	f, err := os.Open("/proc/filesystems")
	if err != nil {
		return err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		if s.Text() == "nodev\toverlay" {
			return nil
		}
	}
	return fmt.Errorf("overlay mount not supported")
}
