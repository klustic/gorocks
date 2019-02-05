package main

import (
	"crypto/tls"
  "flag"
	"log"
	"net"

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

func handleConnection(conn net.Conn, socksHost string) {
	// Wrap connection in yamux
	session, err := yamux.Server(conn, nil)
	if err != nil {
		log.Println("[!] Error initializing yamux tunnel")
		return
	}
	defer conn.Close()
	defer session.Close()

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
