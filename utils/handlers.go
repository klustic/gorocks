package utils

import (
	"io"
	"net"
	"sync"
)

func Proxy(c1 net.Conn, c2 net.Conn) {
	var wg sync.WaitGroup

	intProxy := func(a net.Conn, b net.Conn) {
		defer a.Close()
		defer b.Close()
		io.Copy(a, b)
		wg.Done()
	}

	go intProxy(c1, c2)
	wg.Add(1)

	go intProxy(c2, c1)
	wg.Add(1)

	wg.Wait()

	return
}
