package integration

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/robinmonjo/psdock/notifier"
)

func Test_simpleStart(t *testing.T) {
	beforeTest(t)
	fmt.Printf("testing simple start ... ")
	b := newBinary()
	err := b.start("-image", imagePath, "-rootfs", "/tmp/test_psdock_roo", "ls")
	if err != nil {
		fmt.Println(b.debugInfo())
		t.Fatal(err)
	}
	//TODO check output
	fmt.Println("done")
}

func Test_envAndHostname(t *testing.T) {
	beforeTest(t)
	fmt.Printf("setting env and hostname ... ")
	b := newBinary()
	err := b.start("-image", imagePath, "-rootfs", "/tmp/test_psdock_roo", "-env", "FOO=BAR", "-hostname", "foobar", "bash", "-c", "echo $FOO && hostname")
	if err != nil {
		fmt.Println(b.debugInfo())
		t.Fatal(err)
	}

	cleanStdout := strings.Trim(string(b.stdout), "\n")

	if cleanStdout != "BAR\nfoobar" {
		t.Fatalf("expected output to be BAR\nfoobar got %s", cleanStdout)
	}

	fmt.Println("done")
}

func Test_bindMount(t *testing.T) {
	beforeTest(t)
	fmt.Printf("testing bind mounts ... ")

	//create a temp file
	f, err := ioutil.TempFile("", "psdock_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	b := newBinary()
	test := fmt.Sprintf("test -f %s", f.Name()) // exits 0 if exists, 1 otherwise
	err = b.start("-image", imagePath, "-rootfs", "/tmp/test_psdock_roo", "-bind-mount", "/tmp:/tmp:ro", "bash", "-c", test)
	if err != nil {
		fmt.Println(b.debugInfo())
		t.Fatal(err)
	}

	fmt.Println("done")
}

func Test_webhook(t *testing.T) {
	beforeTest(t)
	fmt.Printf("testing web hooks call ... ")

	cpt := 0
	expectedStatus := []notifier.PsStatus{notifier.StatusStarting, notifier.StatusRunning, notifier.StatusCrashed}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		status := statusFromHookBody(r.Body, t)
		if status != expectedStatus[cpt] {
			t.Fatalf("expecting status %v got %v", expectedStatus[cpt], status)
		}
		cpt++
	}))
	defer ts.Close()

	b := newBinary()
	err := b.start("-image", imagePath, "-rootfs", "/tmp/test_psdock_roo", "-web-hook", ts.URL, "ls")
	if err != nil {
		fmt.Println(b.debugInfo())
		t.Fatal(err)
	}

	if cpt != 3 {
		fmt.Println(b.debugInfo())
		t.Fatalf("hook called %d times, should have been called 3 times", cpt)
	}
	fmt.Println("done")
}

func Test_bindPort(t *testing.T) {
	beforeTest(t)
	fmt.Printf("testing port binding ... ")

	ch := make(chan notifier.PsStatus)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ch <- statusFromHookBody(r.Body, t)
	}))
	defer ts.Close()

	b := newBinary()
	go func() {
		err := b.start("-image", imagePath, "-rootfs", "/tmp/test_psdock_roo", "-web-hook", ts.URL, "-bind-port", "9778", "nc", "-l", "9778")
		if err != nil {
			fmt.Println(b.debugInfo())
			t.Fatal(err)
		}
	}()

	status := <-ch
	if status != notifier.StatusStarting {
		t.Fatalf("expecting status %v got %v", notifier.StatusStarting, status)
	}

	status = <-ch
	if status != notifier.StatusRunning {
		t.Fatalf("expecting status %v got %v", notifier.StatusRunning, status)
	}

	if err := b.stop(); err != nil {
		fmt.Println(b.debugInfo())
		t.Fatal(err)
	}

	status = <-ch
	if status != notifier.StatusCrashed {
		t.Fatalf("expecting status %v got %v", notifier.StatusCrashed, status)
	}

	fmt.Println("done")
}

func Test_remoteStdio(t *testing.T) {
	beforeTest(t)
	fmt.Printf("testing remote stdio ... ")

	//spawn a tcp listener
	ln, err := net.Listen("tcp", ":9999")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	var conn net.Conn
	go func() {
		conn, err = ln.Accept()
		if err != nil {
			t.Fatal(err)
		}
	}()

	//spawn webhook server
	ch := make(chan notifier.PsStatus)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ch <- statusFromHookBody(r.Body, t)
	}))
	defer ts.Close()

	b := newBinary()
	go func() {
		if err := b.start("-image", imagePath, "-rootfs", "/tmp/test_psdock_roo", "-web-hook", ts.URL, "-stdio", "tcp://localhost:9999", "tail", "-f"); err != nil {
			fmt.Println(b.debugInfo())
			t.Fatal(err)
		}
	}()

	status := <-ch
	if status != notifier.StatusStarting {
		t.Fatalf("expecting status %v got %v", notifier.StatusStarting, status)
	}

	status = <-ch
	if status != notifier.StatusRunning {
		t.Fatalf("expecting status %v got %v", notifier.StatusRunning, status)
	}

	//just closing the tcp server should stop the container
	conn.Close()

	status = <-ch
	if status != notifier.StatusCrashed {
		t.Fatalf("expecting status %v got %v", notifier.StatusCrashed, status)
	}
	fmt.Println("done")
}
