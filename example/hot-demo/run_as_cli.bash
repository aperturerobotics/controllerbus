#!/bin/bash
set -eo pipefail
set -x

export CONTROLLER_BUS_CODEGEN_DIR="./codegen-module/"
export CONTROLLER_BUS_PLUGIN_BINARY_ID="controllerbus/examples/hot-demo/codegen-demo/1"
export CONTROLLER_BUS_OUTPUT="./output/demo-plugin.{buildHash}.cbus.so"
export CONTROLLER_BUS_PLUGIN_BUILD_PREFIX="cbus-demo"

mkdir -p ./output
go run -v \
   -trimpath \
   github.com/aperturerobotics/controllerbus/cmd/controllerbus -- \
   hot compile \
  "github.com/aperturerobotics/controllerbus/example/boilerplate/controller" \
  "./demo-controller" \
  "github.com/pkg/errors" \
  "github.com/aperturerobotics/controllerbus/cmd/controllerbus"

