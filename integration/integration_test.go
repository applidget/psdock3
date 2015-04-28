package integration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/robinmonjo/psdock/notifier"
)

func Test_simpleStart(t *testing.T) {
	beforeTest(t)
	fmt.Printf("testing simple start ... ")
	b := newBinary()
	err := b.start("-i", imagePath, "-r", "/tmp/test_psdock_roo", "ls")
	if err != nil {
		t.Fatal(err)
	}
	//TODO check output
	fmt.Println("done")
}

func Test_webhook(t *testing.T) {
	beforeTest(t)
	fmt.Printf("testing web hooks call ... ")
	cpt := 0
	expectedStatus := []notifier.PsStatus{notifier.StatusStarting, notifier.StatusRunning, notifier.StatusCrashed}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var payload notifier.HookPayload

		body, _ := ioutil.ReadAll(r.Body)
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatal(err)
		}
		if payload.Ps.Status != expectedStatus[cpt] {
			t.Fatalf("expecting status %v got %v", expectedStatus[cpt], payload.Ps.Status)
		}
		cpt++
	}))
	defer ts.Close()

	b := newBinary()
	err := b.start("-i", imagePath, "-r", "/tmp/test_psdock_roo", "-wh", ts.URL, "ls")
	if err != nil {
		t.Fatal(err)
	}

	if cpt != 3 {
		t.Fatalf("hook called %d times, should have been called 3 times", cpt)
	}
	fmt.Println("done")
}

func Test_bindPort(t *testing.T) {
	beforeTest(t)
  fmt.Printf("testing port binding ... ")
  cpt := 0
	expectedStatus := []notifier.PsStatus{notifier.StatusStarting, notifier.StatusRunning, notifier.StatusCrashed}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var payload notifier.HookPayload

		body, _ := ioutil.ReadAll(r.Body)
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatal(err)
		}
		if payload.Ps.Status != expectedStatus[cpt] {
			t.Fatalf("expecting status %v got %v", expectedStatus[cpt], payload.Ps.Status)
		}
		cpt++
	}))
	defer ts.Close()

	b := newBinary()
	err := b.start("-i", imagePath, "-r", "/tmp/test_psdock_roo", "-wh", ts.URL, "--bp", "9778", "nc", "-l", "9778")
	if err != nil {
		t.Fatal(err)
	}

	if cpt != 3 {
		t.Fatalf("hook called %d times, should have been called 3 times", cpt)
	}

  fmt.Println("done")
}

func Test_stop(t *testing.T) {
	beforeTest()

}
*/
