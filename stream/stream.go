package stream

import (
	"crypto/tls"
	"io"
	"net"
	"net/url"
	"os"
	"time"
)

const DIAL_TIMEOUT = 5 * time.Second

type Stream struct {
	Input  io.Reader
	Output io.Writer
}

func NewStream(uri string) (*Stream, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	path := u.Host + u.Path

	switch u.Scheme {
	case "":
		return &Stream{os.Stdin, os.Stdout}, nil //use standard input, output

	case "file":
		f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return nil, err
		}
		return &Stream{nil, f}, nil //if stdio is a file, do not support stdin (not intercative)

	case "ssl":
		fallthrough
	case "tls":
		tcpConn, err := net.DialTimeout("tcp", path, DIAL_TIMEOUT)
		if err != nil {
			return nil, err
		}
		config := &tls.Config{InsecureSkipVerify: true}
		conn := tls.Client(tcpConn, config)
		return &Stream{conn, conn}, nil

	default:
		conn, err := net.DialTimeout(u.Scheme, path, DIAL_TIMEOUT)
		if err != nil {
			return nil, err
		}
		return &Stream{conn, conn}, nil
	}
}

func (s *Stream) Close() {
	if rc, ok := s.Input.(io.ReadCloser); ok {
		rc.Close()
	}

	if wc, ok := s.Output.(io.WriteCloser); ok {
		wc.Close()
	}
}
