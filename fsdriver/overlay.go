package fsdriver

import (
	"fmt"
	"os"
	"reflect"
	"syscall"
)

func init() {
	register("overlay", reflect.TypeOf(overlay{}))
}

type overlay struct {
	lowerDir string
	upperDir string
	workDir  string
}

func (o *overlay) Init(image, dest string) error {
	if err := supports("overlay"); err != nil {
		return err
	}
	o.lowerDir = image
	o.upperDir = dest
	o.workDir = fmt.Sprintf("%s_work", dest)

	return nil
}

func (o *overlay) SetupRootfs() error {
	//mount image in readonly into dest
	if err := os.MkdirAll(o.upperDir, 0755); err != nil { // rootfs MUST be with x permission otherwise user switching may fail
		return err
	}
	if err := os.MkdirAll(o.workDir, 0700); err != nil {
		return err
	}
	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", o.lowerDir, o.upperDir, o.workDir)
	return syscall.Mount("overlay", o.upperDir, "overlay", 0, opts)
}

func (o *overlay) CleanupRootfs() error {
	if err := syscall.Unmount(o.upperDir, 0); err != nil {
		return err
	}
	if err := os.RemoveAll(o.upperDir); err != nil {
		return err
	}
	return os.RemoveAll(o.workDir)
}
