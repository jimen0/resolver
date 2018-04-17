package resolver

import (
	"fmt"

	"github.com/miekg/dns"
)

// ZoneTransfer performs a Zone Transfer request.
func ZoneTransfer(host, server string, port int) ([]string, error) {
	fqdn := dns.Fqdn(host)
	msg := new(dns.Msg)
	msg.SetAxfr(fqdn)

	transfer := new(dns.Transfer)
	answers, err := transfer.In(msg, fmt.Sprintf("%s:%d", server, port))
	if err != nil {
		return nil, err
	}

	var resolution []string
	for a := range answers {
		if a.Error != nil {
			continue
		}

		for _, rr := range a.RR {
			var value string
			switch v := rr.(type) {
			case *dns.A:
				value = v.A.String()
			case *dns.CNAME:
				value = v.Target
			case *dns.NS:
				value = v.Ns
			case *dns.PTR:
				value = v.Ptr
			default:
				continue
			}
			resolution = append(resolution, value)
		}
	}

	return resolution, nil
}
