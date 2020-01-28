package ip

import (
	"net"
)

// Our reference type.
type IP struct {
	IPv4 net.IP
	IPv6 net.IP
}

// https://stackoverflow.com/a/48519490/4949938
func (i *IP) IsIPv6Available() bool {
	return i.IPv6.To4() != nil
}
