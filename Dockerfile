FROM golang:1.13.0-alpine as builder
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
WORKDIR /go/src/github.com/mxplusb/cloudflare-sync
COPY go.mod go.sum ./
RUN GO111MODULE=on go mod download
COPY . .
RUN go build -o cloudflare-sync

FROM alpine:3.10
COPY --from=builder /go/src/github.com/mxplusb/cloudflare-sync/cloudflare-sync /cloudflare-sync
CMD /cloudflare-sync \
    -records-file-name=${RECORDS_FILE_NAME} \
    -zone-id=${ZONE_ID} \
    -api-token=${API_TOKEN} \
    -frequency=${FREQUENCY}
