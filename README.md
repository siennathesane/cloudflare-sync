## cloudflare-sync
[![Go Report Card](https://goreportcard.com/badge/github.com/mxplusb/cloudflare-sync)](https://goreportcard.com/report/github.com/mxplusb/cloudflare-sync)

A nice to have, MIT-licensed tool for using Cloudflare as a dynamic DNS provider.

## Usage

Before you get started, ensure that you have a Cloudflare site (one or more, doesn't matter since it's by Zone ID) so records can be updated. Leveraging the `config/example.json`, create a file that you want to contain your own DNS A records. Run `go build -v ./cmd -o cloudflare-sync.exe`, then leverage `cloudflare-sync.exe -h` for the specifics.

Currently this is undergoing an overhaul, so please feel to provide some feedback on its changes.

### Docker

There is a Dockerfile you can use to push to your own registry, if you want. You can also leverage my pre-built one at `cr.r3t.io/library/cloudflare-sync:latest` if you want. Here is a template command you'll want:

```bash
docker run \
    -ti \
    -e "API_TOKEN=''" \
    -e "ZONE_ID=''" \
    -e "FREQUENCY=30" \
    -e "RECORDS_FILE_NAME=production.json" \
    cr.r3t.io/library/cloudflare-sync
```

Don't forget to pass in your own `production.json` file via docker volumes.

### Kubernetes

There is a `kubernetes.yml` file which you can use to deploy a `ConfigMap` and `Deployment` for this. You shouldn't ever need more than one replica. Fill out the `ConfigMap`, `spec.template.spec.containers[0].env` values, and `data.production.json` with your configuration.

## Contributors

This project exists thanks to all the people who contribute. [[Contribute](CONTRIBUTING.md)].
<a href="https://github.com/mxplusb/cloudflare-sync/graphs/contributors"><img src="https://opencollective.com/cloudflare-dyns/contributors.svg?width=890&button=false" /></a>
