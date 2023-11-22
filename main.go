package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/miekg/dns"
)

func main() {
	address := flag.String("a", "127.0.0.1:8053", "specify IPv4 address and port: <address>:<port>. default=127.0.0.1:8053")
	useTls := flag.Bool("tls", false, "indicate to use TLS connection. default=udp")
	flag.Parse()

	log.Printf("Serving DNS queries at %s with TLS:%v\n", *address, *useTls)

	handler := dns.HandlerFunc(handleDns)
	server := serveDns(*address, *useTls, handler)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig

	log.Printf("Signal %s received, stopping server...\n", s)
	server.Shutdown()
}

func handleDns(writer dns.ResponseWriter, m *dns.Msg) {
	log.Println("Received Query:\n", m)
	response, err := resolveDomain(m)

	if err != nil {
		log.Println(err)
		m.MsgHdr.Response = true
		m.MsgHdr.Rcode = 2
		writer.WriteMsg(response)
	} else {
		writer.WriteMsg(response)
	}
}

func serveDns(address string, useTls bool, handler dns.Handler) *dns.Server {
	var server *dns.Server

	if useTls {
		cert, err := tls.LoadX509KeyPair("./dns_server.pem", "./dns_server.key")
		if err != nil {
			log.Fatalln(err)
		}
		cfg := &tls.Config{Certificates: []tls.Certificate{cert}}

		server = &dns.Server{Addr: address, Net: "tcp-tls", Handler: handler, TLSConfig: cfg}
	} else {
		server = &dns.Server{Addr: address, Net: "udp", Handler: handler}
	}

	go server.ListenAndServe()
	return server
}

func resolveDomain(m *dns.Msg) (*dns.Msg, error) {
	client := new(dns.Client)
	destination := GetRandomRootServer()

	for true {
		in, _, err := client.Exchange(m, destination)

		if err != nil {
			return nil, err
		} else if len(in.Answer) > 0 {
			return in, nil
		} else {
			if nsRecord, ok := in.Ns[0].(*dns.NS); ok {
				destination = nsRecord.Ns + ":53"
			}
		}
	}

	return nil, fmt.Errorf("error resolving question %q", m.Question[0])
}
