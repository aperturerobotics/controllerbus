#!/bin/bash
set -eo pipefail
set -x

export GOFLAGS="-buildvcs=false"

go mod tidy
go run -v -trimpath ./ -- daemon
