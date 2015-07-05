package fsdriver

import (
	"os"
	"path/filepath"
)

func createFakeImage() (string, []string, error) {
	//create a fake image
	image := filepath.Join(os.TempDir(), "image_psdock_test")

	if err := os.MkdirAll(image, 0700); err != nil {
		return "", nil, err
	}

	directories := []string{"bin", "boot", "dev", "etc", "home", "lib", "mnt", "opt", "proc", "root", "run", "sbin", "sys", "tmp", "usr", "var"}
	for _, dir := range directories {
		if err := os.MkdirAll(filepath.Join(image, dir), 0700); err != nil {
			return "", nil, err
		}
	}
	return image, directories, nil
}
