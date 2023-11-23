package akdns

import (
	"crypto/tls"
	"fmt"
	"log"

	"github.com/miekg/dns"
)

type TlsClient struct {
	Config *tls.Config
}

func (c *TlsClient) HandleDnsTls(writer dns.ResponseWriter, m *dns.Msg) {
	log.Println("Received Query:\n", m)
	response, err := c.resolveDomainTls(m)

	if err != nil {
		log.Println(err)
		m.MsgHdr.Response = true
		m.MsgHdr.Rcode = 2
		writer.WriteMsg(response)
	} else {
		writer.WriteMsg(response)
	}
}

func (c *TlsClient) ServeDnsTls(address string, handler dns.Handler) (*dns.Server, error) {
	if c.Config == nil {
		return nil, fmt.Errorf("TlsClient does not contain a Config")
	}

	server := &dns.Server{Addr: address, Net: "tcp-tls", Handler: handler, TLSConfig: c.Config}
	go server.ListenAndServe()

	return server, nil
}

func (c *TlsClient) resolveDomainTls(m *dns.Msg) (*dns.Msg, error) {
	destination := getRandomRootServer() + ":853" //853 is designated DoT port

	for true {
		conn, err := dns.DialWithTLS("tcp-tls", destination, c.Config)

		if err != nil {
			return nil, fmt.Errorf("%v", err)
		}

		defer conn.Close()
		conn.WriteMsg(m)
		response, err := conn.ReadMsg()

		if err != nil {
			return nil, err
		} else if len(response.Answer) > 0 {
			return response, nil
		} else {
			if nsRecord, ok := response.Ns[0].(*dns.NS); ok {
				destination = nsRecord.Ns + ":853"
			}
		}
	}

	return nil, fmt.Errorf("error resolving question %q", m.Question[0])
}
