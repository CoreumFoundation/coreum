#!/usr/bin/make -f

build: go.sum
	CGO_ENABLED=1 go build -mod=readonly  -o build/cored ./cmd/cored

install: go.sum
	CGO_ENABLED=1 go install -mod=readonly  ./cmd/cored
