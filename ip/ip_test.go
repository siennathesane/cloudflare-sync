package ip

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIP_IsIPv6Available(t *testing.T) {

	a := assert.New(t)

	i := IP{
		IPv6: net.ParseIP("::FFFF:C0A8:1"),
	}
	res := i.IsIPv6Available()
	a.Equal(true, res, "basic ipv6 should be available")

	i = IP{
		IPv6: net.ParseIP("::FFFF:C0A8:0001"),
	}

	res = i.IsIPv6Available()
	a.Equal(true, res, "ipv6 with leading zeros should be available")

	i = IP{
		IPv6: net.ParseIP("0000:0000:0000:0000:0000:FFFF:C0A8:1"),
	}

	res = i.IsIPv6Available()
	a.Equal(true, res, "double colon expansion ipv6 should be available")

	i = IP{
		IPv6: net.ParseIP("::FFFF:192.168.0.1"),
	}

	res = i.IsIPv6Available()
	a.Equal(true, res, "ipv4 literal should be available")

}
