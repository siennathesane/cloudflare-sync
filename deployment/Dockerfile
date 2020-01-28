FROM golang:alpine

WORKDIR /go/src/cloudflare-sync
COPY .. .
RUN go get -d -v ./...
RUN go install -v ./cmd

CMD cloudflare-sync -records-file-name=${RECORDS_FILE_NAME}
