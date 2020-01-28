package ip

import (
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
	ipm.upstreamRecords, err = ipm.client.DNSRecords(ipm.config.ZoneId, cloudflare.DNSRecord{})
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
	for idx := range ipm.config.Records {
		if payload.IsIPv6Available() && ipm.config.Records[idx].Type == "AAAA" {
			ipm.updateAAAARecord(payload.IPv6, ipm.config.Records[idx])
		}
		if ipm.config.Records[idx].Type == "A" {
			ipm.updateARecord(payload.IPv4, ipm.config.Records[idx])
		}
	}
}

func (ipm *IPManager) updateARecord(ip net.IP, record cloudflare.DNSRecord) {
	record.Content = ip.String()

	for idx := range ipm.upstreamRecords {
		if ipm.upstreamRecords[idx].Name == record.Name {
			record.ID = ipm.upstreamRecords[idx].ID
		}
	}

	ipm.limiter.Take()
	err := ipm.client.UpdateDNSRecord(ipm.config.ZoneId, record.ID, record)
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
	err := ipm.client.UpdateDNSRecord(ipm.config.ZoneId, record.ID, record)
	if err != nil {
		ipm.logger.Printf("error uploading record: %s", err)
		return
	}

	ipm.logger.Printf("updated %s", record.Name)
}
