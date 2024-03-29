#!/bin/bash
set -eo pipefail
set -x

export GOFLAGS="-buildvcs=false"

# export CONTROLLER_BUS_CODEGEN_DIR="./codegen-module/"
export CONTROLLER_BUS_PLUGIN_BINARY_ID="controllerbus/examples/hot-demo"
export CONTROLLER_BUS_OUTPUT="./plugins/demo-plugin.cbus.so"
export CONTROLLER_BUS_PLUGIN_BUILD_PREFIX="cbus-demo"

mkdir -p ./plugins
go run -v \
   -trimpath \
   github.com/aperturerobotics/controllerbus/cmd/controllerbus -- \
   plugin compile \
  "github.com/aperturerobotics/controllerbus/example/boilerplate/controller"
