package main

import (
	"flag"
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

	handler := dns.HandlerFunc(akdns.HandleDnsUdp)
	server := akdns.ServeDnsUdp(*address, handler)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig

	log.Printf("Signal %s received, stopping server...\n", s)
	server.Shutdown()
}
