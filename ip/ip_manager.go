package ip

import (
	"context" // This was added correctly
	"log"
	"net"
	"time"

	"cloudflare-sync/config"
	"github.com/cloudflare/cloudflare-go"
	"go.uber.org/ratelimit"
)

type IPManagerSettings struct {
	Limiter           ratelimit.Limiter
	Config            *config.Config
	Logger            *log.Logger
	BackPressureLimit int
}

type IPManager struct {
	// our settings
	limiter ratelimit.Limiter
	config  *config.Config
	logger  *log.Logger
	client  *cloudflare.API

	// our presets.
	ipQueue     chan IP
	recordQueue chan cloudflare.DNSRecord
	ipify       IIPify

	// discovered
	upstreamRecords []cloudflare.DNSRecord
}

// Start a new IP manager.
func NewIPManager(settings *IPManagerSettings) (*IPManager, error) {
	ipm := &IPManager{
		limiter:     settings.Limiter,
		config:      settings.Config,
		recordQueue: make(chan cloudflare.DNSRecord, settings.BackPressureLimit),
		logger: settings.Logger,
	}

	var err error
	ipm.client, err = settings.Config.NewClient(settings.Logger)
	if err != nil {
		ipm.logger.Printf("error creating cloudflare client: %s", err)
		return &IPManager{}, err
	}

	// try to get the upstream records
	// In cloudflare-go v4.4.0, DNSRecords was replaced by ListDNSRecords.
	// It now requires a context, a ResourceContainer for the zone, and ListDNSRecordsParams.
	// It returns ([]DNSRecord, *ResultInfo, error). We only need the records and the error for now.
	var resultInfo *cloudflare.ResultInfo // Or some other appropriate type if ResultInfo is not what's returned
	ipm.upstreamRecords, resultInfo, err = ipm.client.ListDNSRecords(context.Background(), cloudflare.ZoneIdentifier(ipm.config.ZoneId), cloudflare.ListDNSRecordsParams{})
	// We'll log the resultInfo for now if it's not nil, just to see what it contains.
	if resultInfo != nil {
		ipm.logger.Printf("ListDNSRecords resultInfo: %+v", resultInfo)
	}
	if err != nil {
		ipm.logger.Printf("error fetching upstream records: %s", err)
		return &IPManager{}, err
	}

	// build the ipify implementation
	ipm.ipQueue = make(chan IP, settings.BackPressureLimit)
	ipm.ipify = NewIPify(&IPifySettings{
		Queue:   ipm.ipQueue,
		Limiter: settings.Limiter,
		Logger:  settings.Logger,
	})

	return ipm, nil
}

func (ipm *IPManager) Run() {
	ipm.updateRunner()
	ipm.ticker()
}

func (ipm *IPManager) Die() {
	ipm.logger.Println("cleaning up before dying.")
	close(ipm.ipQueue)
	close(ipm.recordQueue)
	ipm.logger.Println("she's dead, jim.")
}

func r() {
	if r:= recover(); r != nil {
		return
	}
}

// detach the ticker.
func (ipm *IPManager) ticker() {
	go func() {
		defer r()
		ticker := time.NewTicker(time.Duration(ipm.config.Frequency) * time.Second)
		for ; true; <-ticker.C {
			ipm.ipify.GetCurrentAddress()
		}
	}()
}

// this is just to facilitate detaching from the request.
func (ipm *IPManager) updateRunner() {
	go func() {
		for {
			ipm.updateReceiver(<-ipm.ipQueue)
		}
	}()
}

// now we handle the request:wq!
func (ipm *IPManager) updateReceiver(payload IP) {
	// If both IPs are nil (e.g., all IPify providers failed), skip updates.
	if payload.IPv4 == nil && !payload.IsIPv6Available() {
		ipm.logger.Println("Both IPv4 and IPv6 are nil in payload, skipping DNS updates.")
		return
	}

	for idx := range ipm.config.Records {
		// Check for AAAA record update:
		// Only proceed if a valid IPv6 is available in the payload and the record type is AAAA.
		if payload.IsIPv6Available() && ipm.config.Records[idx].Type == "AAAA" {
			ipm.logger.Printf("Attempting to update AAAA record %s with IP %s", ipm.config.Records[idx].Name, payload.IPv6.String())
			ipm.updateAAAARecord(payload.IPv6, ipm.config.Records[idx])
		}
		// Check for A record update:
		// Only proceed if a valid IPv4 is available in the payload and the record type is A.
		if payload.IPv4 != nil && ipm.config.Records[idx].Type == "A" {
			ipm.logger.Printf("Attempting to update A record %s with IP %s", ipm.config.Records[idx].Name, payload.IPv4.String())
			ipm.updateARecord(payload.IPv4, ipm.config.Records[idx])
		}
	}
}

// THIS IS WHERE THE BROKEN UpdateDNSRecord CALLS ARE
func (ipm *IPManager) updateARecord(ip net.IP, record cloudflare.DNSRecord) {
	record.Content = ip.String()

	for idx := range ipm.upstreamRecords {
		if ipm.upstreamRecords[idx].Name == record.Name {
			record.ID = ipm.upstreamRecords[idx].ID
		}
	}

	ipm.limiter.Take()
	// In cloudflare-go v4.4.0, UpdateDNSRecord now requires context, ResourceContainer, and UpdateDNSRecordParams.
	// It returns the updated DNSRecord and an error.
	params := cloudflare.UpdateDNSRecordParams{
		ID:      record.ID,
		Type:    record.Type, // From the matched upstream record / config
		Name:    record.Name, // From the matched upstream record / config
		Content: ip.String(),
		TTL:     record.TTL,     // Preserve TTL from original record spec
		Proxied: record.Proxied, // Preserve proxied status
	}
	// Assign the first return value to a blank identifier if it's not used.
	_, err := ipm.client.UpdateDNSRecord(context.Background(), cloudflare.ZoneIdentifier(ipm.config.ZoneId), params)
	if err != nil {
		ipm.logger.Printf("error uploading record: %s", err)
		return
	}

	ipm.logger.Printf("updated %s.", record.Name)

}

func (ipm *IPManager) updateAAAARecord(ip net.IP, record cloudflare.DNSRecord) {
	record.Content = ip.String()

	for idx := range ipm.upstreamRecords {
		if ipm.upstreamRecords[idx].Name == record.Name {
			record.ID = ipm.upstreamRecords[idx].ID
		}
	}

	ipm.limiter.Take()
	// In cloudflare-go v4.4.0, UpdateDNSRecord now requires context, ResourceContainer, and UpdateDNSRecordParams.
	// It returns the updated DNSRecord and an error.
	params := cloudflare.UpdateDNSRecordParams{
		ID:      record.ID,
		Type:    record.Type,
		Name:    record.Name,
		Content: ip.String(),
		TTL:     record.TTL,
		Proxied: record.Proxied,
	}
	// Assign the first return value to a blank identifier if it's not used.
	_, err := ipm.client.UpdateDNSRecord(context.Background(), cloudflare.ZoneIdentifier(ipm.config.ZoneId), params)
	if err != nil {
		ipm.logger.Printf("error uploading record: %s", err)
		return
	}

	ipm.logger.Printf("updated %s", record.Name)
}
