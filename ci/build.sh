#!/bin/bash

set -e
set -x

go get ./...
mkdir -p build
go build -o build/koinos-util-golang *.go
go build -o build/koinos-util-golang rpc/*.go
