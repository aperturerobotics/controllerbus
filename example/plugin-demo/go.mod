module github.com/aperturerobotics/controllerbus/example/plugin-demo

go 1.27.0

toolchain go1.27.4

replace github.com/aperturerobotics/controllerbus => ../..

require (
	github.com/aperturerobotics/controllerbus v0.53.5 // master
	github.com/aperturerobotics/protobuf-go-lite v0.19.1-0.20261018054529-f60c4d2691ea // latest
	github.com/blang/semver/v4 v4.0.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.10.2
)

require (
	github.com/aperturerobotics/cli v1.1.0 // indirect
	github.com/aperturerobotics/fastjson v0.1.2-0.20260705010846-94f343f5bb34 // indirect
	github.com/aperturerobotics/fsnotify v1.9.1-0.20261007093243-45a75ed654f0 // indirect
	github.com/aperturerobotics/go-websocket v1.8.15-0.20260619192713-a096778f08c1 // indirect
	github.com/aperturerobotics/json-iterator-lite v1.1.0 // indirect
	github.com/aperturerobotics/starpc v0.52.3 // indirect; main
	github.com/aperturerobotics/util v1.34.10-0.20261024061815-104de78e8992 // indirect; latest
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/libp2p/go-yamux/v4 v4.0.2 // indirect
	github.com/libp2p/go-yamux/v5 v5.1.0 // indirect
	github.com/mr-tron/base58 v1.3.0 // indirect
	github.com/xrash/smetrics v0.0.0-20250705151800-55b8f293f342 // indirect
	github.com/zeebo/blake3 v0.2.4 // indirect
	golang.org/x/mod v0.42.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/tools v0.52.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	mvdan.cc/gofumpt v0.12.0 // indirect
)
