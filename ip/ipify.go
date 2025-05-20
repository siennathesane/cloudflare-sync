package ip

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"go.uber.org/ratelimit"
)

var (
	ipv4Providers = []string{
		"https://api.ipify.org",
		"https://ipv4.icanhazip.com",
		"https://myexternalip.com/raw",
	}
	ipv6Providers = []string{"https://api6.ipify.org", "https://ipv6.icanhazip.com"}
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
	// Ensure IPs are nil initially
	ipRef.IPv4 = nil
	ipRef.IPv6 = nil

	// Fetch IPv4
	for _, providerUrl := range ipv4Providers {
		ipy.limiter.Take()
		ipy.logger.Printf("Attempting to fetch IPv4 from %s", providerUrl)
		resp, err := http.Get(providerUrl)
		if err != nil {
			ipy.logger.Printf("Failed to get IPv4 from %s: %v", providerUrl, err)
			continue
		}

		// Ensure body is closed for this attempt
		func() {
			defer func() {
				if closeErr := resp.Body.Close(); closeErr != nil {
					ipy.logger.Printf("Error closing response body from %s: %v", providerUrl, closeErr)
				}
			}()

			if resp.StatusCode != http.StatusOK {
				ipy.logger.Printf("Failed to get IPv4 from %s: status code %d", providerUrl, resp.StatusCode)
				return // continue to next provider via outer loop's continue
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				ipy.logger.Printf("Failed to read body from %s (IPv4): %v", providerUrl, err)
				return // continue
			}

			parsedIP := net.ParseIP(string(body))
			// Ensure it's a valid IPv4 address
			if parsedIP == nil || parsedIP.To4() == nil {
				ipy.logger.Printf("Failed to parse IPv4 address from %s, response: %s", providerUrl, string(body))
				return // continue
			}

			ipRef.IPv4 = parsedIP
			ipy.logger.Printf("Successfully fetched IPv4 %s from %s", ipRef.IPv4, providerUrl)
		}() // End of scope for individual provider response handling

		if ipRef.IPv4 != nil {
			break // Successfully fetched IPv4, exit provider loop
		}
	}

	// Fetch IPv6
	for _, providerUrl := range ipv6Providers {
		ipy.limiter.Take()
		ipy.logger.Printf("Attempting to fetch IPv6 from %s", providerUrl)
		resp, err := http.Get(providerUrl)
		if err != nil {
			ipy.logger.Printf("Failed to get IPv6 from %s: %v", providerUrl, err)
			continue
		}

		func() {
			defer func() {
				if closeErr := resp.Body.Close(); closeErr != nil {
					ipy.logger.Printf("Error closing response body from %s: %v", providerUrl, closeErr)
				}
			}()

			if resp.StatusCode != http.StatusOK {
				ipy.logger.Printf("Failed to get IPv6 from %s: status code %d", providerUrl, resp.StatusCode)
				return // continue
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				ipy.logger.Printf("Failed to read body from %s (IPv6): %v", providerUrl, err)
				return // continue
			}

			parsedIP := net.ParseIP(string(body))
			// Ensure it's a valid IPv6 address (not IPv4-mapped, etc.)
			if parsedIP == nil || parsedIP.To4() != nil || parsedIP.To16() == nil {
				ipy.logger.Printf("Failed to parse IPv6 address from %s, response: %s", providerUrl, string(body))
				return // continue
			} else {
				ipRef.IPv6 = parsedIP
				ipy.logger.Printf("Successfully fetched IPv6 %s from %s", ipRef.IPv6, providerUrl)
			}
		}() // End of scope for individual provider response handling

		if ipRef.IPv6 != nil {
			break // Successfully fetched IPv6, exit provider loop
		}
	}

	ipy.c <- ipRef
}
