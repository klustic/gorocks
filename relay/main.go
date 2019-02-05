package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net"

	socks5 "github.com/armon/go-socks5"
	yamux "github.com/hashicorp/yamux"
)

func connectTunnel(serverHost string) (conn net.Conn, err error) {
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	if conn, err = tls.Dial("tcp", serverHost, conf); err != nil {
		log.Println(err)
		return
	}

	return
}

func main() {
	tunnelAddr := flag.String("tunnel", "127.0.0.1:443", "The host:port address of the server")
	flag.Parse()

	config := yamux.DefaultConfig()

	for {
		// Connect TLS socket wrapped with yamux
		conn, err := connectTunnel(*tunnelAddr)
		if err != nil {
			log.Println("[ERR] Error connecting tunnel: " + err.Error())
			return
		}
		session, err := yamux.Client(conn, config)
		if err != nil {
			log.Println(err)
			return
		}

		// Accept new streams and pass them to a goroutine
		for {
			stream, err := session.Accept()
			if err != nil {
				log.Println("[ERR] Error accept new stream: " + err.Error())
				break
			}

			// Pass to goroutine that handles SOCKS protocol
			conf := &socks5.Config{}
			server, err := socks5.New(conf)
			if err != nil {
				log.Println("[ERR] Error setting up SOCKS server: " + err.Error())
				stream.Close()
				continue
			}
			go server.ServeConn(stream)
		}
	}
}
