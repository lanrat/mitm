package main

import (
	"flag"
	"log"
	"net"
	"strings"

	"github.com/miekg/dns"
)

var (
	listenAddr     = flag.String("listenAddr", "", "address to listen on")
	wildcardAnswer = flag.String("wildcard", "", "wildcard response")
	ntp            = flag.String("ntp", "", "set domains with *ntp* to provided IP")
)

var domainsToAddresses map[string]string = map[string]string{
	"google.com.":                "1.2.3.4",
	"pool.ntp.org.":              "192.168.2.1",
	"api.luxorone.com.":          "192.168.7.2",
	"resident-api.luxerone.com.": "192.168.7.2",
	"apple.com.":                 "192.168.7.2",
}

type handler struct{}

func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)
	//log.Printf("DNS Request: %s", r.String())
	clientIP, _, err := net.SplitHostPort(w.RemoteAddr().String())
	if err != nil {
		log.Fatal(err)
	}
	if len(r.Question) > 1 {
		log.Fatalf("DNS query send multiple questions [%s] %s", clientIP, r.String())
	}
	switch r.Question[0].Qtype {
	case dns.TypeA:
		msg.Authoritative = true
		domain := strings.ToLower(msg.Question[0].Name)
		address, ok := domainsToAddresses[domain]
		if ok {
			msg.Answer = append(msg.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
				A:   net.ParseIP(address),
			})
			log.Printf("DNS Request [%s] %s %s -> %s", clientIP, r.Question[0].Name, dns.TypeToString[r.Question[0].Qtype], address)
		} else {
			switch {
			case len(*ntp) > 0 && strings.Contains(r.Question[0].Name, "ntp"):
				address := *ntp
				msg.Answer = append(msg.Answer, &dns.A{
					Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
					A:   net.ParseIP(address),
				})
				log.Printf("DNS NTP Request [%s] %s %s -> %s", clientIP, r.Question[0].Name, dns.TypeToString[r.Question[0].Qtype], address)
			case len(*wildcardAnswer) > 0:
				address := *wildcardAnswer
				msg.Answer = append(msg.Answer, &dns.A{
					Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
					A:   net.ParseIP(address),
				})
				log.Printf("DNS Wildcard Request [%s] %s %s -> %s", clientIP, r.Question[0].Name, dns.TypeToString[r.Question[0].Qtype], address)
			default:
				log.Printf("DNS Request [%s] %s %s -> ???", clientIP, r.Question[0].Name, dns.TypeToString[r.Question[0].Qtype])
			}
		}
	default:
		log.Printf("DNS Request [%s] %s %s -> ???", clientIP, r.Question[0].Name, dns.TypeToString[r.Question[0].Qtype])
	}
	w.WriteMsg(&msg)
}

func main() {
	flag.Parse()
	listen := net.JoinHostPort(*listenAddr, "53")
	log.Printf("starting dns server on %s", listen)
	srv := &dns.Server{Addr: listen, Net: "udp"}
	srv.Handler = &handler{}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to set udp listener %s\n", err.Error())
	}
}
