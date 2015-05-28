package integration

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/applidget/psdock/notifier"
)

// integration expect to find a ubuntu rootfs in /tmp/image, otherwise they won't be run
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

func statusFromHookBody(body io.Reader, t *testing.T) notifier.PsStatus {
	var payload notifier.HookPayload

	content, _ := ioutil.ReadAll(body)
	if err := json.Unmarshal(content, &payload); err != nil {
		t.Fatal(err)
	}
	return payload.Ps.Status
}
