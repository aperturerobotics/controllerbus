module github.com/aperturerobotics/controllerbus/example/plugin-demo

go 1.27.0

toolchain go1.27.1

replace github.com/aperturerobotics/controllerbus => ../..

require (
	github.com/aperturerobotics/controllerbus v0.53.5 // master
	github.com/aperturerobotics/protobuf-go-lite v0.18.0 // latest
	github.com/blang/semver/v4 v4.0.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.10.2
)

require (
	github.com/aperturerobotics/cli v1.1.0 // indirect
	github.com/aperturerobotics/fastjson v0.1.1 // indirect
	github.com/aperturerobotics/fsnotify v1.9.1-0.20260506231828-931cb4bf1761 // indirect
	github.com/aperturerobotics/go-websocket v1.8.15-0.20260619192713-a096778f08c1 // indirect
	github.com/aperturerobotics/json-iterator-lite v1.1.0 // indirect
	github.com/aperturerobotics/starpc v0.52.1 // indirect; main
	github.com/aperturerobotics/util v1.34.10-0.20260802062101-496aab6cefd2 // indirect; latest
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/libp2p/go-yamux/v4 v4.0.2 // indirect
	github.com/libp2p/go-yamux/v5 v5.1.0 // indirect
	github.com/mr-tron/base58 v1.3.0 // indirect
	github.com/xrash/smetrics v0.0.0-20250705151800-55b8f293f342 // indirect
	github.com/zeebo/blake3 v0.2.4 // indirect
	golang.org/x/mod v0.41.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/tools v0.50.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	mvdan.cc/gofumpt v0.12.0 // indirect
)
