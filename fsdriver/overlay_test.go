package fsdriver

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func Test_overlay(t *testing.T) {
	for i := 0; i < 5; i++ { //looping to test "robustness"
		fmt.Printf("overlay rootfs ... ")
		//create a fake image
		image := path.Join(os.TempDir(), "image_psdock_test")

		if err := os.MkdirAll(image, 0700); err != nil {
			t.Fatal(err)
		}

		directories := []string{"bin", "boot", "dev", "etc", "home", "lib", "mnt", "opt", "proc", "root", "run", "sbin", "sys", "tmp", "usr", "var"}
		for _, dir := range directories {
			if err := os.MkdirAll(path.Join(image, dir), 0700); err != nil {
				t.Fatal(err)
			}
		}

		rootfs := path.Join(os.TempDir(), "rootfs_psdock_test")
		o, err := NewOverlay(image, rootfs)
		if err != nil {
			t.Fatal(err)
		}

		if err := o.SetupRootfs(); err != nil {
			t.Fatal(err)
		}
		defer o.CleanupRootfs()

		//check rootfs is the same as image
		mountedDirectories, err := ioutil.ReadDir(o.upperDir)
		if len(mountedDirectories) != len(directories) {
			t.Fatalf("%d mounted directories expected %d", len(mountedDirectories), len(directories))
		}

		for i, dir := range mountedDirectories {
			if dir.Name() != directories[i] {
				t.Fatalf("mounted directory mismatch, expecting %s, got %s", directories[i], dir.Name())
			}
		}

		if err := o.CleanupRootfs(); err != nil {
			t.Fatal(err)
		}

		if _, err := os.Stat(rootfs); err == nil {
			t.Fatalf("rootfs %s not properly cleaned up", rootfs)
		}

		if _, err := os.Stat(o.workDir); err == nil {
			t.Fatalf("rootfs work dir %s not properly cleaned up", o.workDir)
		}

		fmt.Println("done")
	}
}
