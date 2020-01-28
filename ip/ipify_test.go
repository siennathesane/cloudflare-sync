package ip

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/ratelimit"
)

var settings = &IPifySettings{
	Queue:   make(chan IP, 1),
	Limiter: ratelimit.New(1),
	Logger: log.New(os.Stdout, "", log.LstdFlags),
}

func TestNewIPify(t *testing.T) {
	n := NewIPify(settings)
	assert.NotNil(t, n, "builder should not be nil")
}

func TestIPify_GetCurrentAddress(t *testing.T) {

	a := assert.New(t)

	n := NewIPify(settings)
	n.GetCurrentAddress()

	res := <- settings.Queue
	a.NotEmpty(res.IPv4, "the ipv4 address should not be empty")

	if res.IsIPv6Available() {
		a.NotEmpty(res.IPv6, "the ipv6 address should not be empty")
	}
}
