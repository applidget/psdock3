package fsdriver

import (
	"fmt"
	"os"
	"reflect"
	"syscall"
)

func init() {
	register("aufs", reflect.TypeOf(aufs{}))
}

type aufs struct {
	lowerDir string
	upperDir string
}

func (a *aufs) Init(image, dest string) error {
	if err := supports("aufs"); err != nil {
		return err
	}
	a.lowerDir = image
	a.upperDir = dest

	return nil
}

func (a *aufs) SetupRootfs() error {
	//mount image in readonly into dest
	if err := os.MkdirAll(a.upperDir, 0755); err != nil {
		return err
	}
	opts := fmt.Sprintf("br=%s=rw:%s=ro", a.upperDir, a.lowerDir)
	return syscall.Mount("aufs", a.upperDir, "aufs", 0, opts)
}

func (a *aufs) CleanupRootfs() error {
	if err := syscall.Unmount(a.upperDir, 0); err != nil {
		return err
	}
	return os.RemoveAll(a.upperDir)
}
