package resolver

import "testing"

func TestReverse(t *testing.T) {
	tt := []struct {
		name string
		ip   string
		exp  string
	}{
		{
			name: "valid IPv4",
			ip:   "8.8.8.8",
			exp:  "8.8.8.8.in-addr.arpa.",
		},
		{
			name: "valid IPv6",
			ip:   "2001:4860:4860::8888",
			exp:  "",
		},
		{
			name: "invalid address",
			ip:   "gopher",
			exp:  "",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			r, err := reverse(tc.ip)
			if err != nil && tc.exp != "" {
				t.Fatal(err)
			}

			if tc.exp != r {
				t.Fatalf("expected %s got %s", tc.exp, r)
			}
		})
	}
}
