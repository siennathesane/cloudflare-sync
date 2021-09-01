## cloudflare-sync
[![Go Report Card](https://goreportcard.com/badge/github.com/mxplusb/cloudflare-sync)](https://goreportcard.com/report/github.com/mxplusb/cloudflare-sync)

A nice to have, MIT-licensed tool for using Cloudflare as a dynamic DNS provider.

### Project Status (1 Sept 2021)

This project was originally archived because I didn't have the time or desire to develop in my personal life for a long time, and I didn't want to lead folks on that I would support this. I've decided to unarchive this project because lots of people have shown interest in it through more stars. I'll work on getting some CI and other things put into place so it's easier to accept changes in the coming weeks and months. This is a useful tool and I do like that folks are interested in using it more.

Maybe in time it can support more than just Cloudflare, but for now, I want to make sure that integration is rock solid for lots of different DNS record types.

If you want to help maintain this, let me know, I'm definitely open to involving others, even though it's a very small, very basic program. 🙂

## Usage

Before you get started, ensure that you have a Cloudflare site (one or more, doesn't matter since it's by Zone ID) so records can be updated. Leveraging the `config/example.json`, create a file that you want to contain your own DNS A records. Run `go build -v ./cmd -o cloudflare-sync.exe`, then leverage `cloudflare-sync.exe -h` for the specifics.

Currently this is undergoing an overhaul, so please feel to provide some feedback on its changes.

### Configuration

The configuration is file-based, you can find an example of the schema in `config/example.json`. You need to pass this file via a flag, which you can find with `cloudflare-sync.exe --help`. This file is a subset of the Cloudflare API, so if you don't want it to override the values you already have in Cloudflare, just make sure they match. Please feel free to update this configuration section with more commentary if my explanation isn't satisfactory.

### Docker

There is a Dockerfile you can use to push to your own registry, if you want. Here is a template command you'll want:

```bash
docker run \
    -ti \
    -e "API_TOKEN=''" \
    -e "ZONE_ID=''" \
    -e "FREQUENCY=30" \
    -e "RECORDS_FILE_NAME=production.json" \
    <your-image>
```

Don't forget to pass in your own `production.json` file via docker volumes.

### Kubernetes

There is a `kubernetes.yml` file which you can use to deploy a `ConfigMap` and `Deployment` for this. You shouldn't ever need more than one replica. Fill out the `ConfigMap`, `spec.template.spec.containers[0].env` values, and `data.production.json` with your configuration.

## Contributors

This project exists thanks to all the people who contribute. [[Contribute](CONTRIBUTING.md)].
<a href="https://github.com/mxplusb/cloudflare-sync/graphs/contributors"><img src="https://opencollective.com/cloudflare-dyns/contributors.svg?width=890&button=false" /></a>
