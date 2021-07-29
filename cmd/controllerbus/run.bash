#!/bin/bash
set -eo pipefail
set -x

go mod tidy
go run -v -trimpath ./ -- daemon
