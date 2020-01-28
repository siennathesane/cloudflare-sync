package ip

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"go.uber.org/ratelimit"
)

// How do we want to interact with IPify?
type IIPify interface {
	GetCurrentAddress()
}

// How we are interacting with IPify
type IPify struct {
	c       chan IP
	logger  *log.Logger
	limiter ratelimit.Limiter
}

// Our settings.
type IPifySettings struct {
	Queue chan IP
	Limiter ratelimit.Limiter
	Logger *log.Logger
}

// Build a new IPify implementation
func NewIPify(settings *IPifySettings) *IPify {
	return &IPify{
		c: settings.Queue,
		limiter: settings.Limiter,
		logger: settings.Logger,
	}
}

func (ipy *IPify) GetCurrentAddress() {

	ipy.logger.Println("refreshing public ip.")

	var ipRef IP

	ipy.limiter.Take()
	resp4, err := http.Get("https://api.ipify.org")
	if err != nil {
		ipy.logger.Fatalf("cannot get ip: %s", err)
		return
	}

	if resp4.StatusCode != http.StatusOK {
		ipy.logger.Printf("cannot read response from ipify, response code: %d", resp4.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(resp4.Body)
	if err != nil {
		ipy.logger.Fatalf("cannot read ipify response: %s", err)
		return
	}

	ipRef.IPv4 = net.ParseIP(string(body))
	ipy.logger.Printf("current public ipv4 is %s.", ipRef.IPv4)

	if err := resp4.Body.Close(); err != nil {
		ipy.logger.Fatal(err)
		return
	}

	ipy.limiter.Take()
	resp6, err := http.Get("https://api6.ipify.org")
	if err != nil {
		ipy.logger.Printf("cannot get ip: %s", err)
		return
	}

	if resp6.StatusCode != http.StatusOK {
		ipy.logger.Printf("cannot read response from ipify, response code: %d", resp6.StatusCode)
	}

	body, err = ioutil.ReadAll(resp6.Body)
	if err != nil {
		ipy.logger.Fatalf("cannot read ipify response: %s", err)
		return
	}

	ipRef.IPv6 = net.ParseIP(string(body))

	if ipRef.IsIPv6Available() {
		ipy.logger.Printf("current public ipv6 is %s.", ipRef.IPv6)
	}

	if err := resp4.Body.Close(); err != nil {
		ipy.logger.Fatal(err)
		return
	}

	ipy.c <- ipRef
}
