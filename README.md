## cloudflare-sync
[![Go Report Card](https://goreportcard.com/badge/github.com/mxplusb/cloudflare-sync)](https://goreportcard.com/report/github.com/mxplusb/cloudflare-sync)
[![Financial Contributors on Open Collective](https://opencollective.com/cloudflare-dyns/all/badge.svg?label=financial+contributors)](https://opencollective.com/cloudflare-dyns) 

A nice to have, MIT-licensed tool for using Cloudflare as a dynamic DNS provider.

## Usage

Before you get started, ensure that you have a Cloudflare site (one or more, doesn't matter since it's by Zone ID) so records can be updated. Leveraging the `example.json`, create a file that you want to contain your own DNS A records. Run `go build -v .`, then leverage `cloudflare-sync.exe -h` for the specifics.

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

### Code Contributors

This project exists thanks to all the people who contribute. [[Contribute](CONTRIBUTING.md)].
<a href="https://github.com/mxplusb/cloudflare-sync/graphs/contributors"><img src="https://opencollective.com/cloudflare-dyns/contributors.svg?width=890&button=false" /></a>

### Financial Contributors

Become a financial contributor and help us sustain our community. [[Contribute](https://opencollective.com/cloudflare-dyns/contribute)]

#### Individuals

<a href="https://opencollective.com/cloudflare-dyns"><img src="https://opencollective.com/cloudflare-dyns/individuals.svg?width=890"></a>

#### Organizations

Support this project with your organization. Your logo will show up here with a link to your website. [[Contribute](https://opencollective.com/cloudflare-dyns/contribute)]

<a href="https://opencollective.com/cloudflare-dyns/organization/0/website"><img src="https://opencollective.com/cloudflare-dyns/organization/0/avatar.svg"></a>
<a href="https://opencollective.com/cloudflare-dyns/organization/1/website"><img src="https://opencollective.com/cloudflare-dyns/organization/1/avatar.svg"></a>
<a href="https://opencollective.com/cloudflare-dyns/organization/2/website"><img src="https://opencollective.com/cloudflare-dyns/organization/2/avatar.svg"></a>
<a href="https://opencollective.com/cloudflare-dyns/organization/3/website"><img src="https://opencollective.com/cloudflare-dyns/organization/3/avatar.svg"></a>
<a href="https://opencollective.com/cloudflare-dyns/organization/4/website"><img src="https://opencollective.com/cloudflare-dyns/organization/4/avatar.svg"></a>
<a href="https://opencollective.com/cloudflare-dyns/organization/5/website"><img src="https://opencollective.com/cloudflare-dyns/organization/5/avatar.svg"></a>
<a href="https://opencollective.com/cloudflare-dyns/organization/6/website"><img src="https://opencollective.com/cloudflare-dyns/organization/6/avatar.svg"></a>
<a href="https://opencollective.com/cloudflare-dyns/organization/7/website"><img src="https://opencollective.com/cloudflare-dyns/organization/7/avatar.svg"></a>
<a href="https://opencollective.com/cloudflare-dyns/organization/8/website"><img src="https://opencollective.com/cloudflare-dyns/organization/8/avatar.svg"></a>
<a href="https://opencollective.com/cloudflare-dyns/organization/9/website"><img src="https://opencollective.com/cloudflare-dyns/organization/9/avatar.svg"></a>
