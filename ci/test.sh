#!/bin/bash

set -e
set -x

go test -coverprofile=./build/util.out ./...

gcov2lcov -infile=./build/util.out -outfile=./build/util.info

golangci-lint run ./...
