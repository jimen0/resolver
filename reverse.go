package resolver

import (
	"errors"
	"net"
	"strconv"
)

// reverse returns the in-addr.arpa. hostname of the IP address suitable
// for rDNS (PTR) record lookup or an error if it fails to to parse the IP
// address.
func reverse(s string) (string, error) {
	ip := net.ParseIP(s)
	if ip == nil {
		return "", errors.New("invalid IP address")
	}

	if ip.To4() == nil {
		return "", errors.New("non IPv4 address")
	}

	return strconv.Itoa(int(ip[15])) + "." + strconv.Itoa(int(ip[14])) + "." + strconv.Itoa(int(ip[13])) + "." +
		strconv.Itoa(int(ip[12])) + ".in-addr.arpa.", nil
}
