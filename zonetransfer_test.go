package resolver

import (
	"testing"

	"github.com/jimen0/submassive/collector/helper"
)

func TestZoneTransfer(t *testing.T) {
	tt := []struct {
		name string
		host string
		srv  string
		port int
		exp  []string
	}{
		{
			name: "valid transfer",
			host: "zonetransfer.me",
			srv:  "nsztm1.digi.ninja.",
			port: 53,
			exp:  []string{"5.196.105.14", "nsztm1.digi.ninja.", "nsztm2.digi.ninja.", "www.zonetransfer.me.", "127.0.0.1", "202.14.81.230", "143.228.181.132", "74.125.206.26", "127.0.0.1", "intns1.zonetransfer.me.", "intns2.zonetransfer.me.", "81.4.108.41", "167.88.42.94", "4.23.39.254", "207.46.197.32", "www.sydneyoperahouse.com.", "127.0.0.1", "www.zonetransfer.me.", "174.36.59.154", "5.196.105.14"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			dom, err := ZoneTransfer(tc.host, tc.srv, tc.port)
			if err != nil && tc.exp != nil {
				t.Fatal(err)
			}

			if !helper.Equal(t, tc.exp, dom) {
				t.Fatalf("expected %v got %v", tc.exp, dom)
			}
		})
	}
}
