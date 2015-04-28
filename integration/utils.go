package integration

import (
	"fmt"
	"os"
	"testing"
)

const imagePath = "/tmp/image"

var serverRunning bool

// integration expect to find a ubuntu rootfs in /tmp/image, other wise they won't be run
func beforeTest(t *testing.T) {
	if !fileExists(imagePath) {
		fmt.Printf("skipping, ubuntu image not found in %s\n", imagePath)
		t.Skip()
	}
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return false
	}
	return true
}
