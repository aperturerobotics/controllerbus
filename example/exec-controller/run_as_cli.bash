#!/bin/bash
set -eo pipefail

go run -v \
   -trimpath \
   github.com/aperturerobotics/controllerbus/cmd/controllerbus -- \
   daemon --config="" &
PID=$?

function cleanup {
    kill -9 $PID
}
trap cleanup EXIT

sleep 1

go run -v \
   -trimpath \
   github.com/aperturerobotics/controllerbus/cmd/controllerbus -- \
   client exec -f exec-controller.yaml
