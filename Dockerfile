FROM golang:alpine

WORKDIR /go/src/cloudflare-sync
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...

CMD cloudflare-sync \
    -records-file-name=${RECORDS_FILE_NAME} \
    -zone-id=${ZONE_ID} \
    -api-token=${API_TOKEN} \
    -frequency=${FREQUENCY}
