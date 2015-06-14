package fsdriver

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"reflect"
)

// Driver create a usable rootfs from an imutable image directory
type Driver interface {
	Init(image, rootfs string) error
	SetupRootfs() error
	CleanupRootfs() error
}

var (
	drivers        []string = []string{"overlay", "aufs"}
	driverRegistry          = make(map[string]reflect.Type)
)

func register(driverName string, driverType reflect.Type) {
	driverRegistry[driverName] = driverType
}

// Return a driver, in order it returns overlay and if not supported, return aufs.
func New(image, rootfs string) (Driver, error) {
	for _, name := range drivers {
		v := reflect.New(driverRegistry[name])
		d, ok := v.Interface().(Driver)
		if !ok {
			return nil, fmt.Errorf("%s driver doesn't seem to implement the psdock.Driver interface")
		}

		if err := d.Init(image, rootfs); err == nil {
			return d, nil
		}
	}

	return nil, fmt.Errorf("none of %v drivers are supported on the host", drivers)
}

func supports(name string) error {
	exec.Command("modprobe", name).Run()

	f, err := os.Open("/proc/filesystems")
	if err != nil {
		return err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		if s.Text() == "nodev\t"+name {
			return nil
		}
	}
	return fmt.Errorf("%s mount not supported", name)
}
