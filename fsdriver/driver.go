package driver

import ()

// Driver create a usable rootfs from an imutable image directory
type Driver interface {
	SetupRootfs(image string, dest string) error
}
