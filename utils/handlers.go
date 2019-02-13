package utils

import (
	"io"
	"net"
	"sync"
)

func Proxy(c1 net.Conn, c2 net.Conn) {
	defer c1.Close()
	defer c2.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		io.Copy(c1, c2)
		wg.Done()
	}()

	go func() {
		io.Copy(c2, c1)
		wg.Done()
	}()

	wg.Wait()

	return
}
