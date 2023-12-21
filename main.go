package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
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
	var tlsClient *akdns.TlsClient
	handler := dns.HandlerFunc(akdns.HandleDnsUdp)
	udpServer := akdns.ServeDnsUdp(*address, handler)

	if *tlsAddress != "" {
		var err error
		tlsServer, tlsClient, err = serverDnsTls(*tlsAddress)

		if err != nil {
			log.Fatalf("%v\n", err)
		}
	}

	handleShutdown(udpServer, tlsServer, tlsClient)
}

func serverDnsTls(address string) (*dns.Server, *akdns.TlsClient, error) {
	config, err := loadTlsConfig("./dns_server.pem", "./dns_server.key")
	if err != nil {
		return nil, nil, err
	}

	tlsClient := akdns.TlsClient{
		Config: config,
		Pool:   make(map[string]*dns.Conn),
		Cache:  &akdns.RecordCache{Cache: &sync.Map{}},
	}

	handler := dns.HandlerFunc(tlsClient.HandleDnsTls)
	tlsServer, err := tlsClient.ServeDnsTls(address, handler)

	if err != nil {
		return nil, nil, err
	}

	return tlsServer, &tlsClient, nil
}

func loadTlsConfig(publicKey, privateKey string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(publicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	config := &tls.Config{Certificates: []tls.Certificate{cert}}

	return config, nil
}

func handleShutdown(udpServer, tlsServer *dns.Server, tlsClient *akdns.TlsClient) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig

	log.Printf("Signal %s received, stopping server...\n", s)

	udpServer.Shutdown()

	if tlsServer != nil {
		tlsServer.Shutdown()
		errors := tlsClient.CloseConnectionPools()

		if len(errors) > 0 {
			for _, err := range errors {
				log.Printf("%v\n", err)
			}
		}
	}

	log.Printf("Done!\n")
}
