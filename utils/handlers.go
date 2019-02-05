package utils

import (
	"net"
)

// chanFromConn creates a channel from a Conn object, and sends everything it
// Read()s from the socket to the channel.
// Credit: https://www.stavros.io/posts/proxying-two-connections-go/
func chanFromConn(conn net.Conn) chan []byte {
	c := make(chan []byte)
	go func() {
		b := make([]byte, 1024)
		for {
			n, err := conn.Read(b)
			if n > 0 {
				res := make([]byte, n)
				copy(res, b[:n])
				c <- res
			}
			if err != nil {
				c <- nil
				break
			}
		}
	}()
	return c
}

// Pipe creates a full-duplex pipe between the two sockets and transfers data from one to the other.
// Credit: https://www.stavros.io/posts/proxying-two-connections-go/
func Pipe(conn1 net.Conn, conn2 net.Conn) (err error) {

	chan1 := chanFromConn(conn1)
	chan2 := chanFromConn(conn2)

	defer conn1.Close()
	defer conn2.Close()

	for {
		select {
		case b1 := <-chan1:
			if b1 == nil {
				return
			} else {
				if _, err = conn2.Write(b1); err != nil {
					return
				}
			}
		case b2 := <-chan2:
			if b2 == nil {
				return
			} else {
				if _, err = conn1.Write(b2); err != nil {
					return
				}
			}
		}
	}
}

func PerformProxy(localSock net.Conn, stream net.Conn) (err error) {
	defer stream.Close()
	defer localSock.Close()
	err = Pipe(localSock, stream)
	return
}
