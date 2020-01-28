package config

import (
	"errors"
	"log"

	"github.com/cloudflare/cloudflare-go"
)

type Config struct {
	ZoneId    string                 `json:"zone_id"`
	ZoneName  string                 `json:"zone_name"`
	ApiToken  string                 `json:"api_token"`
	Frequency int                    `json:"frequency"`
	Records   []cloudflare.DNSRecord `json:"records"`
	client    *cloudflare.API
}

func (c *Config) NewClient(logger *log.Logger) (*cloudflare.API, error) {
	var err error

	if !c.Validate() {
		return &cloudflare.API{}, errors.New("invalid c")
	}

	c.client, err = cloudflare.NewWithAPIToken(c.ApiToken)
	if err != nil {
		logger.Fatalf("cannot instantiate cloudflare client: %s", err)
	}

	if c.ZoneId == "" && c.ZoneName != "" {
		c.ZoneId, err = c.client.ZoneIDByName(c.ZoneName)
		if err != nil {
			logger.Fatalf("error getting zoneId from zoneName: %s", err)
		} else {
			logger.Printf("zoneName %s resolved into zoneId: %s", c.ZoneName, c.ZoneId)
		}
	} else if c.ZoneId != "" && c.ZoneName != "" {
		logger.Printf("zoneName %s will be ignored since a zoneId is setted", c.ZoneName)
	}

	return c.client, nil
}

func (c *Config) Validate() bool {

	okayConfig := true

	if c.ApiToken == "" {
		okayConfig = false
	}

	return okayConfig
}
