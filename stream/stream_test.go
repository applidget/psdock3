package stream

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"testing"
)

func Test_tlsStream(t *testing.T) {
	fmt.Printf("TLS/SSL stream ... ")

	fmt.Println("done")
}

func Test_tcpStream(t *testing.T) {
	fmt.Printf("TCP stream ... ")

	fmt.Println("done")
}

func Test_fileStream(t *testing.T) {
	fmt.Printf("File stream ... ")
	s, err := NewStream("file:///tmp/psdock_test.log")
	if err != nil {
		t.Fatal(err)
	}
	if s.Input != nil {
		t.Fatal("file stream are not expected to have input streams")
	}
	s.Output.Write([]byte("foo bar"))
	s.Close()

	content, err := ioutil.ReadFile("/tmp/psdock_test.log")
	if err != nil {
		t.Fatal(err)
	}
	os.Remove("/tmp/psdock_test.log")
	if string(content) != "foo bar" {
		t.Fatalf("expecting \"foo bar\" got %s", string(content))
	}

	fmt.Println("done")
}
