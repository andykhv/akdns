package main

import (
	"fmt"
	"github.com/miekg/dns"
)

func main() {
    handler := dns.HandlerFunc(ServeDNS)
	err := dns.ListenAndServe("127.0.0.1:8053", "udp", handler)

	if err != nil {
		fmt.Println(err)
	}
}

func ServeDNS(writer dns.ResponseWriter, m *dns.Msg) {
    response, err := resolveDomain(m)

    if err != nil {
        fmt.Println(err)
        m.MsgHdr.Response = true
        m.MsgHdr.Rcode = 2
        writer.WriteMsg(response)
    } else {
        writer.WriteMsg(response)
    }
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

