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
	address := flag.String("addr", "127.0.0.1:8053", "Specify IPv4 address and port: <address>:<port>. default=127.0.0.1:8053")
	tlsAddress := flag.String("tls", "", "Indicate address for tcp. If nothing specified, a TLS port is not opened.")
	flag.Parse()

	if *tlsAddress != "" {
		log.Printf("Serving DNS queries at %s with TLS at %v\n", *address, *tlsAddress)
	} else {
		log.Printf("Serving DNS queries at %s\n", *address)
	}
	var tlsServer *dns.Server

	handler := dns.HandlerFunc(akdns.HandleDnsUdp)
	udpServer := akdns.ServeDnsUdp(*address, handler)

	if *tlsAddress != "" {
		config, err := loadTlsConfig("./dns_server.pem", "./dns_server.key")
		if err != nil {
			log.Fatalln(err)
		}

		tlsClient := akdns.TlsClient{
			Config: config,
		}

		handler := dns.HandlerFunc(tlsClient.HandleDnsTls)
		tlsServer, err = tlsClient.ServeDnsTls(*tlsAddress, handler)

		if err != nil {
			log.Fatalln(err)
		}
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig

	log.Printf("Signal %s received, stopping server...\n", s)

	udpServer.Shutdown()

	if *tlsAddress != "" {
		tlsServer.Shutdown()
	}

	log.Printf("Done!\n")
}

func loadTlsConfig(publicKey, privateKey string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(publicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	config := &tls.Config{Certificates: []tls.Certificate{cert}}

	return config, nil
}
