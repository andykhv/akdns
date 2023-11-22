package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/miekg/dns"
)

func main() {
	log.Printf("Serving DNS queries @localhost:8053\n")
	handler := dns.HandlerFunc(handleDns)
	server := serveDns("127.0.0.1:8053", "udp", handler)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig
	log.Printf("Signal (%s) received, stopping server...\n", s)
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

func serveDns(address, net string, handler dns.Handler) *dns.Server {
	server := &dns.Server{Addr: address, Net: net, Handler: handler}
	go server.ListenAndServe()
	return server
}

func resolveDomain(m *dns.Msg) (*dns.Msg, error) {
	client := new(dns.Client)
	destination := "198.41.0.4:53" //a.root-servers.net

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
