#!/bin/bash

set -e
set -x

go test -v github.com/koinos/koinos-util -coverprofile=./build/util.out -coverpkg=./...

gcov2lcov -infile=./build/util.out -outfile=./build/util.info

golint -set_exit_status ./...
