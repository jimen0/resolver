package resolver

import (
	"context"
	"reflect"
	"testing"

	"github.com/jimen0/submassive/collector/helper"
)

var (
	defaultRetries = 3
	defaultServers = []string{"8.8.8.8:53"}
)

func TestNew(t *testing.T) {
	tt := []struct {
		name    string
		record  string
		retries int
		workers int
		exp     *Resolver
	}{
		{
			name:    "valid",
			record:  "A",
			retries: defaultRetries,
			workers: 1,
			exp:     &Resolver{Record: "A", Retries: defaultRetries, Workers: 1},
		},
		{
			name:    "invalid retries",
			record:  "A",
			retries: -1,
			workers: 1,
			exp:     nil,
		},
		{
			name:    "invalid workers",
			record:  "A",
			retries: defaultRetries,
			workers: 0,
			exp:     nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			out, err := New(tc.record, tc.retries, tc.workers)
			if err != nil && tc.exp != nil {
				t.Fatalf("expected %v got %v", tc.exp, out)
			}
		})
	}
}

func TestResolveList(t *testing.T) {
	tt := []struct {
		name    string
		r       *Resolver
		domains []string
		servers []string
		exp     []result
	}{
		{
			name: "A record 1 worker valid query",
			r: &Resolver{
				Workers: 1,
				Record:  "A",
				Retries: 0,
			},
			domains: []string{"scanme.nmap.org"},
			servers: defaultServers,
			exp:     []result{result{name: "scanme.nmap.org", destination: []string{"45.33.32.156"}}},
		},
		{
			name: "no DNS servers",
			r: &Resolver{
				Workers: 1,
				Record:  "A",
				Retries: 0,
			},
			domains: []string{"scanme.nmap.org"},
			servers: []string{},
			exp:     nil,
		},
		{
			name: "more workers than domains",
			r: &Resolver{
				Workers: 2,
				Record:  "CNAME",
				Retries: 1,
			},
			domains: []string{"hub.github.com"},
			servers: defaultServers,
			exp:     []result{result{name: "hub.github.com", destination: []string{"github.map.fastly.net."}}},
		},
		{
			name: "resolve using non existing DNS server",
			r: &Resolver{
				Workers: 1,
				Record:  "CNAME",
				Retries: 1,
			},
			domains: []string{"hub.github.com"},
			servers: []string{},
			exp:     nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			out := make(chan result)
			done := make(chan struct{})
			var res []result

			go func() {
				for v := range out {
					res = append(res, v)
				}
				done <- struct{}{}
			}()

			err := tc.r.ResolveList(context.Background(), tc.domains, tc.servers, out)
			if err != nil {
				if err != nil && tc.exp != nil {
					t.Fatalf("failed to resolve %v using %v and record %s: %v", tc.domains, tc.servers, tc.r.Record, err)
				}
			}
			<-done

			if !reflect.DeepEqual(tc.exp, res) {
				t.Fatalf("expected %v got %v", tc.exp, out)
			}
		})
	}
}

func TestResolve(t *testing.T) {
	tt := []struct {
		name    string
		host    string
		record  string
		retries int
		srv     []string
		exp     []string
	}{
		{
			name:   "valid A record",
			host:   "scanme.nmap.org",
			srv:    defaultServers,
			record: "A",
			exp:    []string{"45.33.32.156"},
		},
		{
			name:   "valid CNAME record",
			host:   "hub.github.com",
			srv:    defaultServers,
			record: "CNAME",
			exp:    []string{"github.map.fastly.net."},
		},
		{
			name:   "valid PTR record",
			host:   "8.8.8.8",
			srv:    defaultServers,
			record: "PTR",
			exp:    []string{"google-public-dns-a.google.com."},
		},
		{
			name:   "invalid domain",
			host:   "notvalid",
			srv:    defaultServers,
			record: "A",
			exp:    nil,
		},
		{
			name:   "no valid DNS servers",
			host:   "scanme.nmap.org",
			srv:    []string{},
			record: "A",
			exp:    nil,
		},
		{
			name:    "one valid and one invalid DNS server",
			host:    "hub.github.com",
			srv:     []string{"127.0.0.1:0", "8.8.8.8:53"},
			record:  "CNAME",
			retries: 1,
			exp:     []string{"github.map.fastly.net."},
		},
		{
			name:    "multiple invalid DNS servers",
			host:    "hub.github.com",
			srv:     []string{"127.0.0.1:0", "127.0.0.1:0"},
			record:  "CNAME",
			retries: 1,
			exp:     nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			out, err := Resolve(context.Background(), tc.record, tc.host, tc.retries, tc.srv)
			if err != nil && tc.exp != nil {
				t.Fatalf("failed to get record %s for %s using %v: %v", tc.record, tc.host, tc.srv, err)
			}

			if !helper.Equal(t, out, tc.exp) {
				t.Fatalf("want %v got %v", tc.exp, out)
			}
		})
	}
}
