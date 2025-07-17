#!/bin/bash
set -eo pipefail

# change to git root directory
cd "$(git rev-parse --show-toplevel)"

# delete old output dir
rm -rf ./goscript/@goscript

go run github.com/aperturerobotics/goscript/cmd/goscript -- \
  compile \
    --all-deps \
    --output ./goscript \
    --package github.com/aperturerobotics/controllerbus/goscript/target

# --package github.com/aperturerobotics/hydra/unixfs
