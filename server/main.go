package main

import (
	"crypto/tls"
  "flag"
	"fmt"
	"log"
	"net"
	"os"

	yamux "github.com/hashicorp/yamux"
	utils "github.com/klustic/gorocks/utils"
)

func createTunnelServer(cert string, key string, host string) (server net.Listener, err error) {
	cer, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		log.Println(err)
		return
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	server, err = tls.Listen("tcp", host, config)
	if err != nil {
		log.Println(err)
		return
	}

	return
}

func HandleStats(session *yamux.Session, statPath string) {
	// Set up the Unix socket
	l, err := net.Listen("unix", statPath)
	if err != nil {
		fmt.Println("[ERR] Error listening on Unix socket: " + err.Error())
		return
	}
	defer l.Close()
	defer os.Remove(statPath)

	// Wait for connections on the socket and serve stats
	for {
		fd, err := l.Accept()
		if err != nil {
			fmt.Println("[ERR] Error accepting on Unix socket: " + err.Error())
				return
		}

		NumStreams := session.NumStreams()
		rtt, _ := session.Ping()
		msg := fmt.Sprintf("[GOROCKS Stats]\n---\n- Remote Tunnel Peer: %v\n- Round Trip Time: %v\n- Number of Streams: %d\n---\n", session.RemoteAddr(), rtt, NumStreams)

		// Send the log message and close the file
		fd.Write([]byte(msg))
		fd.Close()
	}
	return
}

func handleConnection(conn net.Conn, socksHost string) {
	// Wrap connection in yamux
	session, err := yamux.Server(conn, nil)
	if err != nil {
		log.Println("[!] Error initializing yamux tunnel")
		return
	}
	defer conn.Close()
	defer session.Close()

	// Run the statistics goroutine
	go HandleStats(session, "/tmp/gorocks.sock")

	// Listen on SOCKSv5 server port
	socksServer, err := net.Listen("tcp", socksHost)
	if err != nil {
		log.Println("[!] Error listening for SOCKSv5 clients: " + err.Error())
		return
	}
	defer socksServer.Close()

	// Accept on socksv5 port; open new stream and start new goroutine to proxy
	log.Println("[+] Waiting for SOCKS clients on " + socksHost)
	for {
		client, err := socksServer.Accept()
		if err != nil {
			log.Println("[!] Error accepting connection from SOCKS client: " + err.Error())
			continue
		}
		stream, err := session.Open()
		if err != nil {
			log.Println("[+] Error opening new stream in tunnel: " + err.Error())
			return
		}
		go utils.PerformProxy(client, stream)
	}
	return
}

func main() {
  tunnelAddress := flag.String("tunnel", "0.0.0.0:443", "The bind address on which to accept tunnel connections")
  socksAddress := flag.String("socks", "127.0.0.1:1080", "The bind address on which to accept SOCKSv5 clients")
  certFile := flag.String("cert", "server.crt", "A file containing a TLS certificate")
  keyFile := flag.String("key", "server.key", "A file containing a TLS key")
  flag.Parse()

	server, err := createTunnelServer(*certFile, *keyFile, *tunnelAddress)
	if err != nil {
		log.Println(err)
		return
	}
	defer server.Close()

	for {
		log.Println("Waiting for connections...")
		conn, err := server.Accept()
		if err != nil {
			log.Println(err)
			return
		}
		handleConnection(conn, *socksAddress)
	}
}
