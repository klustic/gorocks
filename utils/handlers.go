package utils

import (
	"io"
	"net"
)

type closeWriter interface {
	CloseWrite() error
}

func Proxy(dst net.Conn, src net.Conn) error {
	// Modified from github.com/armon/go-socks5/request.go
	defer dst.Close()
	defer src.Close()

	errCh := make(chan error, 2)

	_proxy := func(a io.Writer, b io.Reader) {
		_, err := io.Copy(a, b)
		if tcpConn, ok := a.(closeWriter); ok {
			tcpConn.CloseWrite()
		}
		errCh <- err
	}

	// Start proxying
	go _proxy(dst, src)
	go _proxy(src, dst)

	// Wait
	for i := 0; i < 2; i++ {
		e := <-errCh
		if e != nil {
			return e
		}
	}

	return nil
}
