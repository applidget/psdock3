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
	if err := os.MkdirAll(a.upperDir, 0700); err != nil {
		return err
	}
	//res = system("sudo mount -t aufs -o br=#{cont_path}/wlayer=rw:#{base_cont_path}=ro none #{cont_path}/image")
	opts := fmt.Sprintf("br=%s=rw:%s=ro", a.lowerDir, a.upperDir)
	//Mount(source string, target string, fstype string, flags uintptr, data string)
	//syscall.Mount("overlay", o.upperDir, "overlay", 0, opts)
	return syscall.Mount("aufs", a.upperDir, "aufs", 0, opts)
}

func (a *aufs) CleanupRootfs() error {
	if err := syscall.Unmount(a.upperDir, 0); err != nil {
		return err
	}
	return os.RemoveAll(a.upperDir)
}
