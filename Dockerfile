FROM golang:1.16-alpine

# go get uses git to fetch modules
RUN apk add --no-cache git

WORKDIR /media/muafafif/data/goprojects/github.com/afif0808/testkumparan/app

ENV GO111MODULE=on

RUN go mod download
