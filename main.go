package main

import (
	"fmt"
	"github.com/miekg/dns"
)

func main() {
	client := new(dns.Client)
    domain := "facebook.com."
	destination := "198.41.0.4:53" //a.root-servers.net
    var finalMessage **dns.Msg
    var finalError *error

	for finalMessage == nil  && finalError == nil {
        message := newQuestion(domain)
		in, _, err := client.Exchange(message, destination)

        if err != nil {
            finalError = &err
        } else if len(in.Answer) > 0 {
            finalMessage = &in
		} else {
			if nsRecord, ok := in.Ns[0].(*dns.NS); ok {
				destination = nsRecord.Ns + ":53"
			}
		}
	}

    if finalError != nil {
        fmt.Println(*finalError)
    } else {
        fmt.Println(*finalMessage)
    }
}

func newQuestion(domain string) *dns.Msg {
	message := new(dns.Msg)
	message.Id = dns.Id()
	message.Question = make([]dns.Question, 1)
	message.Question[0] = dns.Question{domain, dns.TypeA, dns.ClassINET}
	message.RecursionDesired = true
    message.SetEdns0(4096, false)

    return message
}
