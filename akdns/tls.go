package akdns

import (
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"github.com/miekg/dns"
)

const (
	CLOUDFLARE_DOT_SERVER = "1.1.1.1:853"
	DOT_PORT              = ":853"
)

type TlsClient struct {
	Config *tls.Config
	Pool   map[string]*dns.Conn
	Cache  *RecordCache
}

func (c *TlsClient) HandleDnsTls(writer dns.ResponseWriter, m *dns.Msg) {
	log.Printf("Received Query:\n%v\n", m)
	response, err := c.resolveDomainTls(m)

	if err != nil {
		log.Println(err)
		m.MsgHdr.Response = true
		m.MsgHdr.Rcode = 2
		writer.WriteMsg(m)
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
	destination := CLOUDFLARE_DOT_SERVER

	if records, found := c.Cache.LoadRecord(RecordKey{m.Question[0].Name, m.Question[0].Qtype, m.Question[0].Qclass}); found {
		m.Answer = records
		m.MsgHdr.Response = true
		return m, nil
	}

	for {
		conn, err := c.getConnection(destination)

		if err != nil {
			return nil, fmt.Errorf("%v", err)
		}

		conn.WriteMsg(m)
		response, err := conn.ReadMsg()

		if err != nil {
			return nil, err
		} else if len(response.Answer) > 0 {
			recordKey := RecordKey{m.Question[0].Name, m.Question[0].Qtype, m.Question[0].Qclass}
			ttl := int64(response.Answer[0].Header().Ttl)
			var second int64 = 1000000000
			c.Cache.StoreRecord(recordKey, response.Answer, time.Duration(ttl*second))
			return response, nil
		} else {
			if nsRecord, ok := response.Ns[0].(*dns.NS); ok {
				destination = nsRecord.Ns + DOT_PORT
			}
		}
	}
}

func (c *TlsClient) getConnection(address string) (*dns.Conn, error) {
	conn, found := c.Pool[address]

	if found {
		_, err := conn.Read(make([]byte, 0, 1))

		if err == nil {
			return conn, nil
		}
	}

	conn, err := dns.DialWithTLS("tcp-tls", address, c.Config)

	if err != nil {
		return nil, err
	}

	c.Pool[address] = conn
	return conn, nil
}

func (c *TlsClient) CloseConnectionPools() []error {
	errs := make([]error, 0, 10)

	for _, conn := range c.Pool {
		err := conn.Close()

		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
