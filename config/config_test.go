package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/stretchr/testify/assert"
)

func TestConfig_NewClient(t *testing.T) {

	a := assert.New(t)

	// grab the test file.
	p, _ := filepath.Abs("..")
	fp, _ := filepath.Abs(path.Join(p, "testdata", "production-test.json"))

	b, err := ioutil.ReadFile(fp)
	a.NoError(err, "there should be no errors loading the test json")

	var config Config
	err = json.Unmarshal(b, &config)
	a.NoError(err, "there should be no errors when parsing the test json")
	a.Equal(config.Records[0].Name, "test.r3t.io", "the test record name should be `test.r3t.io`")

	client, err := config.NewClient(log.New(os.Stdout, "", log.LstdFlags))
	a.NoError(err, "there should be no errors when creating a new client")

	records, err := client.DNSRecords(config.ZoneId, cloudflare.DNSRecord{})
	a.NoError(err, "there should be no errors when reading records")
	a.NotEmpty(records, "the records should not be empty")
}
