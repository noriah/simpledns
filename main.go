package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/miekg/dns"
)

const defaultTTL = 300

type generatedZone = []dns.RR

type Config struct {
	ListenAddress string `json:"listen"`
	Zones         []Zone `json:"zones"`
}

type Zone struct {
	Name    string      `json:"name"`
	Records []ZoneEntry `json:"records"`
}

type ZoneEntry struct {
	RecordType string `json:"type"`
	Value      string `json:"value"`
}

func genHdr(zoneName string, rrType uint16) dns.RR_Header {
	return dns.RR_Header{
		Name:   zoneName,
		Rrtype: rrType,
		Class:  dns.ClassINET,
		Ttl:    defaultTTL,
	}
}

func genRecord(zoneName string, entry ZoneEntry) dns.RR {
	switch entry.RecordType {
	case "a", "A":
		return &dns.A{
			Hdr: genHdr(zoneName, dns.TypeA),
			A:   net.ParseIP(entry.Value),
		}
	case "aaaa", "AAAA":
		return &dns.AAAA{
			Hdr:  genHdr(zoneName, dns.TypeAAAA),
			AAAA: net.ParseIP(entry.Value),
		}
	case "text", "TEXT", "txt", "TXT":
		return &dns.TXT{
			Hdr: genHdr(zoneName, dns.TypeTXT),
			Txt: []string{entry.Value},
		}
	case "mx", "MX":
		return &dns.MX{
			Hdr:        genHdr(zoneName, dns.TypeMX),
			Mx:         entry.Value,
			Preference: 10,
		}
	}

	panic(fmt.Sprintf("bad type for %s: %s", zoneName, entry.RecordType))
}

func generateZone(zone Zone) generatedZone {
	outZone := make([]dns.RR, 0)

	for _, entry := range zone.Records {
		record := genRecord(zone.Name, entry)
		outZone = append(outZone, record)
		log.Printf("added %s\n", record)
	}

	return outZone
}

func nameHandler(zone generatedZone) dns.HandlerFunc {
	return func(w dns.ResponseWriter, r *dns.Msg) {
		reply := dns.Msg{}
		reply.SetReply(r)

		for _, q := range r.Question {
			for _, rr := range zone {
				if q.Qtype == rr.Header().Rrtype || q.Qtype == dns.TypeANY {
					reply.Answer = append(reply.Answer, rr)
					log.Printf("serving %s\n", rr)
				}
			}
		}

		w.WriteMsg(&reply)
	}
}

// registers handlers for all Zones in config, then starts a dns server on the
// ListenAddress in config
func dnsServer(config *Config) error {
	for _, e := range config.Zones {
		dns.HandleFunc(e.Name, nameHandler(generateZone(e)))
	}

	addr, err := net.ResolveUDPAddr("udp", config.ListenAddress)
	if err != nil {
		return err
	}

	socket, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	return dns.ActivateAndServe(nil, socket, dns.DefaultServeMux)
}

// reads the configuration from `path`
func readConfig(path string) (config Config, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &config)

	return
}

// runs splitdns
func main() {
	if len(os.Args) < 2 {
		log.Fatalf("path to config required\n")
	}

	config, err := readConfig(os.Args[1])

	if err != nil {
		log.Fatalf("config read failed: %s\n", err)
	}

	if err := dnsServer(&config); err != nil {
		log.Fatalf("dns server quit with non-zero: %s\n", err)
	}
}
