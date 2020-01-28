package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"

	"cloudflare-sync/config"
	"cloudflare-sync/ip"
	"go.uber.org/ratelimit"
)

var (
	filePath string
)

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
	flag.StringVar(&filePath, "records-file-name", "production.json", "Path to the production.json file.")

	flag.Usage = Usage
}

func main() {
	run()
}

func run() {

	flag.Parse()
	log.Println("hello from boulder.")

	// build the new ip manager
	ipm, err := ip.NewIPManager(&ip.IPManagerSettings{
		Limiter: ratelimit.New(4, ratelimit.WithoutSlack),
		Config: func() *config.Config {
			b, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.Fatalf("cannot find production configuration file at %s", filePath)
			}
			var c config.Config
			err = json.Unmarshal(b, &c)
			if err != nil {
				log.Fatalf("cannot unmarshal production configuration file")
			}
			return &c
		}(),
		Logger:            log.New(os.Stdout, "", log.LstdFlags),
		BackPressureLimit: 100,
	})
	if err != nil {
		log.Fatalf("error instantiating ip manager: %s", err)
	}

	ipm.Run()

	quit := make(chan struct{})
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			ipm.Die()
			log.Println("all done.")
			os.Exit(0)
		}
	}()
	<-quit
}
