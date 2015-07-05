package fsdriver

//commented as I don't run tests on a aufs ready box, however this has been tested

/*import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func Test_aufs(t *testing.T) {
	for i := 0; i < 5; i++ { //looping to test "robustness"
		fmt.Printf("aufs rootfs ... ")
		//create a fake image
		image, directories, err := createFakeImage()
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(image)

		rootfs := path.Join(os.TempDir(), "rootfs_psdock_test")
		a := &aufs{}
		if err := a.Init(image, rootfs); err != nil {
			t.Fatal(err)
		}

		if err := a.SetupRootfs(); err != nil {
			t.Fatal(err)
		}
		defer a.CleanupRootfs()

		//check rootfs is the same as image
		mountedDirectories, _ := ioutil.ReadDir(a.upperDir)
		if len(mountedDirectories) != len(directories) {
			t.Fatalf("%d mounted directories expected %d", len(mountedDirectories), len(directories))
		}

		for i, dir := range mountedDirectories {
			if dir.Name() != directories[i] {
				t.Fatalf("mounted directory mismatch, expecting %s, got %s", directories[i], dir.Name())
			}
		}

		if err := a.CleanupRootfs(); err != nil {
			t.Fatal(err)
		}

		if _, err := os.Stat(rootfs); err == nil {
			t.Fatalf("rootfs %s not properly cleaned up", rootfs)
		}

		fmt.Println("done")
	}
}
*/
