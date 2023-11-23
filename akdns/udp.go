package akdns

import (
	"fmt"
	"log"

	"github.com/miekg/dns"
)

func HandleDnsUdp(writer dns.ResponseWriter, m *dns.Msg) {
	log.Println("Received Query:\n", m)
	response, err := resolveDomainUdp(m)

	if err != nil {
		log.Println(err)
		m.MsgHdr.Response = true
		m.MsgHdr.Rcode = 2
		writer.WriteMsg(response)
	} else {
		writer.WriteMsg(response)
	}
}

func resolveDomainUdp(m *dns.Msg) (*dns.Msg, error) {
	client := new(dns.Client)
	destination := getRandomRootServer()

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

func ServeDnsUdp(address string, handler dns.Handler) *dns.Server {
	server := &dns.Server{Addr: address, Net: "udp", Handler: handler}
	go server.ListenAndServe()
	return server
}
