package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"go.uber.org/ratelimit"
)

var (
	filePath        string
	internalRecords []cloudflare.DNSRecord
	upstreamRecords []cloudflare.DNSRecord
	zoneId          string
	zoneName        string
	apiToken        string
	frequency       int
	limiter         ratelimit.Limiter
)

type IP struct {
	IP string
}

type IPify struct {
	client *cloudflare.API
	c      chan IP
	log    *log.Logger
}

var Usage = func() {

	var s []string
	switch runtime.GOOS {
	case "windows":
		s = strings.Split(os.Args[0], `\`)
	default:
		s = strings.Split(os.Args[0], "/")
	}

	fmt.Fprintf(os.Stderr, "\nUse Cloudflare as a dynamic DNS provider.\n\n"+
		"Arguments of %s:\n", s[len(s)-1])

	flag.PrintDefaults()
}

func init() {
	flag.StringVar(&filePath, "records-file-name", "production.json", "Path to the "+
		"production.json file.")
	flag.StringVar(&zoneId, "zone-id", "", "ID of the zone in Cloudflare.")
	flag.StringVar(&zoneName, "zone-name", "", "Name of the zone in Cloudflare.")
	flag.StringVar(&apiToken, "api-token", "", "Cloudflare API token.")
	flag.IntVar(&frequency, "frequency", 30, "Frequency in seconds to update the records. Will "+
		"respect Cloudflare's rate limit, regardless of how many records are configured.")

	flag.Usage = Usage
}

func main() {

	flag.Parse()

	logger := log.New(os.Stdout, "", log.LstdFlags)

	logger.Println("hello from boulder.")

	limiter = ratelimit.New(4, ratelimit.WithoutSlack) // cloudflare's rate limit.

	ipNotifier := make(chan IP, 10)
	client, err := cloudflare.NewWithAPIToken(apiToken)
	if err != nil {
		logger.Fatalf("cannot instantiate cloudflare client: %s", err)
	}

	if zoneId == "" && zoneName != "" {
		zoneId, err = client.ZoneIDByName(zoneName)
		if err != nil {
			logger.Fatalf("error getting zoneId from zoneName: %s", err)
		} else {
			logger.Printf("zoneName %s resolved into zoneId: %s", zoneName, zoneId)
		}
	} else if zoneId != "" && zoneName != "" {
		logger.Printf("zoneName %s will be ignored since a zoneId is setted", zoneName)
	}

	fh, err := ioutil.ReadFile(filePath)
	if err != nil {
		logger.Fatalf("error reading reference file: %s", err)
	}

	if err := json.Unmarshal(fh, &internalRecords); err != nil {
		logger.Fatalf("cannot marshal json from reference file: %s", err)
	}

	ipy := &IPify{
		client: client,
		c:      ipNotifier,
		log:    logger,
	}

	logger.Println("starting up workers.")

	go ipy.findIPAddress()
	go ipy.updateCloudflare()

	logger.Println("workers booted.")

	quit := make(chan struct{})
	<- quit
}

func (ipy *IPify) findIPAddress() {
	ticker := time.NewTicker(time.Second * time.Duration(frequency))

	for _ = range ticker.C {

		ipy.log.Println("refreshing public ip.")

		resp, err := http.Get("https://api.ipify.org?format=json")
		if err != nil {
			ipy.log.Fatalf("cannot get ip: %s", err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			ipy.log.Fatalf("cannot read ipify response: %s", err)
		}

		var ip IP
		if err := json.Unmarshal(body, &ip); err != nil {
			ipy.log.Fatal(err)
		}

		ipy.log.Printf("current public ip is %s.", ip.IP)

		if err := resp.Body.Close(); err != nil {
			ipy.log.Fatal(err)
		}

		limiter.Take()
		upstreamRecords, err = ipy.client.DNSRecords(zoneId, cloudflare.DNSRecord{})
		if err != nil {
			ipy.log.Fatalf("cannot get upstream records: %s", err)
		}

		ipy.c <- ip

		ipy.log.Println("sent record update.")
	}
}

func (ipy *IPify) updateCloudflare() {
	for update := range ipy.c {

		go func() {

			ipy.log.Println("pushing cloudflare update.")

			for idx, _ := range internalRecords {

				internalRecords[idx].Content = update.IP

				for nidx := range upstreamRecords {
					if internalRecords[idx].Name == upstreamRecords[nidx].Name {
						internalRecords[idx].ID = upstreamRecords[nidx].ID
					}
				}

				ipy.log.Printf("updating %s.", internalRecords[idx].Name)

				limiter.Take()
				if err := ipy.client.UpdateDNSRecord(zoneId, internalRecords[idx].ID, internalRecords[idx]); err != nil {
					ipy.log.Fatalf("cannot update cloudflare dns record %s: %s", internalRecords[idx].Name, err)
				}

				ipy.log.Printf("updated %s.", internalRecords[idx].Name)
			}
		}()

		ipy.log.Println("cloudflare update pushed.")
	}
}
