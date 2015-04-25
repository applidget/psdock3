package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
)

type psStatus string

const (
	statusStarting psStatus = "starting"
	statusRunning  psStatus = "running"
	statusCrashed  psStatus = "crashed"
)

var webHook string

func notifyHook(status psStatus) error {
	body := []byte(fmt.Sprintf(`{"ps": { "status": "%s" } }`, string(status)))

	req, err := http.NewRequest("PUT", webHook, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	u, err := url.Parse(webHook)
	if err != nil {
		return err
	}

	var client *http.Client

	if u.Scheme == "https" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	} else {
		client = &http.Client{}
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("bad status code expected 200 .. 299 got %d", resp.Status)
	}
	return nil
}
