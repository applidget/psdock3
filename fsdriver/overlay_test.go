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
		image, directories, err := createFakeImage()
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(image)

		rootfs := path.Join(os.TempDir(), "rootfs_psdock_test")
		o := &overlay{}
		if err := o.Init(image, rootfs); err != nil {
			t.Fatal(err)
		}

		if err := o.SetupRootfs(); err != nil {
			t.Fatal(err)
		}
		defer o.CleanupRootfs()

		//check rootfs is the same as image
		mountedDirectories, _ := ioutil.ReadDir(o.upperDir)
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
