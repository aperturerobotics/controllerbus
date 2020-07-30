#!/bin/bash
set -eo pipefail
set -x

go run -v \
   -trimpath \
   github.com/aperturerobotics/controllerbus/cmd/controllerbus -- \
   client exec -f exec-controller.yaml

