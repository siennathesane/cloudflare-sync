package ip

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"cloudflare-sync/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/ratelimit"
)

var (
	// grab the test file.
	conf = func() *config.Config {
		p, _ := filepath.Abs("..")
		fp, _ := filepath.Abs(path.Join(p, "testdata", "production-test.json"))
		b, _ := ioutil.ReadFile(fp)
		var config config.Config
		_ = json.Unmarshal(b, &config)
		return &config
	}()
)

func TestNewIPManager(t *testing.T) {
	a := assert.New(t)

	t.Logf("%v", conf)

	ipm, err := NewIPManager(&IPManagerSettings{
		Limiter:           ratelimit.New(1),
		Config:            conf,
		Logger:            log.New(os.Stdout, "", log.LstdFlags),
		BackPressureLimit: 10,
	})
	a.NoError(err, "there should not be an error when creating the ip manager")
	a.NotNil(ipm, "the ip manager should not be nil")
}

func TestIPManager_ticker(t *testing.T) {
	a := assert.New(t)

	t.Logf("%v", conf)

	ipm, err := NewIPManager(&IPManagerSettings{
		Limiter:           ratelimit.New(1),
		Config:            conf,
		Logger:            log.New(os.Stdout, "", log.LstdFlags),
		BackPressureLimit: 10,
	})
	a.NoError(err, "there should not be an error when creating the ip manager")
	a.NotNil(ipm, "the ip manager should not be nil")

	ipm.ticker()
	resp := <- ipm.ipQueue
	ipm.Die()

	a.NotNil(resp.IPv4, "the ipv4 address should not be nil")
}

func TestIPManager_Run(t *testing.T) {

	if testing.Short() {
		t.Skip()
	}

	a := assert.New(t)

	t.Logf("%v", conf)

	ipm, err := NewIPManager(&IPManagerSettings{
		Limiter:           ratelimit.New(1),
		Config:            conf,
		Logger:            log.New(os.Stdout, "", log.LstdFlags),
		BackPressureLimit: 10,
	})
	a.NoError(err, "there should not be an error when creating the ip manager")
	a.NotNil(ipm, "the ip manager should not be nil")

	timer := time.NewTimer(90 * time.Second)

	ipm.Run()

	<-timer.C

	ipm.Die()
}
