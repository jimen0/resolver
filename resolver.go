package resolver

import (
	"context"
	"errors"
	"sync"

	"github.com/miekg/dns"
)

// result represents a domain name and its destinations.
type result struct {
	// name is the domain name.
	name string
	// destination is the list of addresses where it resolves to.
	destination []string
}

// ErrNoResponse is returned when all resolution retries were done and no
// response has been received yet.
var ErrNoResponse = errors.New("no response")

// Resolver represents a DNS resolver.
type Resolver struct {
	// Record is the DNS record that is queried.
	Record string
	// Retries is the number of retries done.
	Retries int
	// Workers is the number of concurrent goroutines used during resolution.
	Workers int
}

// New creates a DNS resolver.
func New(record string, retries, workers int) (*Resolver, error) {
	if retries < 0 {
		return nil, errors.New("number of retries must be higher or equals to 0")
	}
	if workers <= 0 {
		return nil, errors.New("at least one worker is needed")
	}

	return &Resolver{
		Record:  record,
		Retries: retries,
		Workers: workers,
	}, nil
}

// ResolveList resolves a slice of hosts and returns the destintations
// they resolve to over the out channel.
func (r *Resolver) ResolveList(ctx context.Context, domains, servers []string, out chan<- result) error {
	defer close(out)

	if len(domains) == 0 {
		return errors.New("at least one domain is needed")
	}

	if len(servers) == 0 {
		return errors.New("at least one DNS server is needed")
	}

	if r.Workers > len(domains) {
		r.Workers = len(domains)
	}

	errs := make(chan error)
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(r.Workers)

	chans := make([]chan string, r.Workers)
	for i := 0; i < len(chans); i++ {
		chans[i] = make(chan string)
	}

	for _, ch := range chans {
		go func(c chan string) {
			defer wg.Done()

			select {
			case <-done:
				return
			default: // avoid blocking
			}

			for v := range c {
				// a server from servers slice should be picked in a way that load is
				// balanced between them all.
				var count int
				srv := []string{servers[count%len(servers)]}
				for i := 0; i < r.Retries; i++ {
					srv = append(srv, servers[(count+i)%len(servers)])
					count++
				}

				dst, err := Resolve(ctx, r.Record, v, r.Retries, srv)
				if err != nil {
					if err != ErrNoResponse {
						errs <- err
					}
					continue
				}

				if len(dst) > 0 {
					out <- result{name: v, destination: dst}
				}
			}
		}(ch)
	}

	for k, v := range domains {
		select {
		case err := <-errs:
			done <- struct{}{}
			return err
		default: // avoid blocking
		}
		chans[k%r.Workers] <- v
	}

	for _, c := range chans {
		close(c)
	}

	wg.Wait()
	return nil

}

// Resolve tries to resolve a host using the given DNS servers.
// If all servers fail to resolve it, Resolve returns an error
func Resolve(ctx context.Context, record, host string, retries int, srv []string) ([]string, error) {
	if len(srv) == 0 {
		return nil, errors.New("at least one DNS server is needed")
	}

	msg := new(dns.Msg)
	msg.Id = dns.Id()
	msg.RecursionDesired = false

	switch record {
	case "A":
		msg.SetQuestion(dns.Fqdn(host), dns.TypeA)
	case "CNAME":
		msg.SetQuestion(dns.Fqdn(host), dns.TypeCNAME)
	case "PTR":
		h, err := reverse(host)
		if err != nil {
			return nil, err
		}
		msg.SetQuestion(dns.Fqdn(h), dns.TypePTR)
	case "NS":
		msg.SetQuestion(dns.Fqdn(host), dns.TypeNS)
	default:
		return nil, errors.New("invalid record")
	}

	var in *dns.Msg
	var err error
	in, err = dns.ExchangeContext(ctx, msg, srv[0])
	for err != nil && retries > 0 {
		in, err = dns.ExchangeContext(ctx, msg, srv[retries%len(srv)])
		if in != nil {
			break
		}
		retries--
	}

	// if no response was received, return an error.
	if in == nil {
		return nil, ErrNoResponse
	}

	var resolution []string
	for _, rr := range in.Answer {
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

	return resolution, nil
}
