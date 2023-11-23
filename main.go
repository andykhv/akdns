package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/andykhv/akdns/akdns"
	"github.com/miekg/dns"
)

func main() {
	address := flag.String("a", "127.0.0.1:8053", "specify IPv4 address and port: <address>:<port>. default=127.0.0.1:8053")
	useTls := flag.Bool("tls", false, "indicate to use TLS connection. default=udp")
	flag.Parse()

	log.Printf("Serving DNS queries at %s with TLS:%v\n", *address, *useTls)
	var server *dns.Server

	if !*useTls {
		handler := dns.HandlerFunc(akdns.HandleDnsUdp)
		server = akdns.ServeDnsUdp(*address, handler)
	} else {
		config, err := loadTlsConfig("./dns_server.pem", "./dns_server.key")
		if err != nil {
			log.Fatalln(err)
		}

		tlsClient := akdns.TlsClient{
			Config: config,
		}

		handler := dns.HandlerFunc(tlsClient.HandleDnsTls)
		server, err = tlsClient.ServeDnsTls(*address, handler)

		if err != nil {
			log.Fatalln(err)
		}
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig

	log.Printf("Signal %s received, stopping server...\n", s)
	server.Shutdown()
}

func loadTlsConfig(publicKey, privateKey string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(publicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	config := &tls.Config{Certificates: []tls.Certificate{cert}}

	return config, nil
}
