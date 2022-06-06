FROM golang:1.18-alpine as builder
RUN apk add --no-cache git make gcc libc-dev linux-headers
RUN set -eux; apk add --no-cache ca-certificates build-base

WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go mod tidy
RUN BUILD_TAGS=muslc go install ./cmd/coremon

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /go/bin/* /usr/local/bin/
WORKDIR /apps/data

ENTRYPOINT [ "coremon" ]
