module github.com/aperturerobotics/controllerbus/example/plugin-demo

go 1.25.0

toolchain go1.26.2

replace github.com/aperturerobotics/controllerbus => ../..

require (
	github.com/aperturerobotics/controllerbus v0.53.2 // master
	github.com/aperturerobotics/protobuf-go-lite v0.13.0 // latest
	github.com/aperturerobotics/starpc v0.49.7 // indirect; main
	github.com/aperturerobotics/util v1.34.3 // indirect; latest
)

require (
	github.com/blang/semver/v4 v4.0.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.5-0.20260426203557-6878cb36b029
)

require (
	github.com/aperturerobotics/cli v1.1.0 // indirect
	github.com/aperturerobotics/fsnotify v1.9.1-0.20260329111252-827e5e9feeab // indirect
	github.com/aperturerobotics/go-websocket v1.8.15-0.20260329113544-74dbfb8f11c6 // indirect
	github.com/aperturerobotics/json-iterator-lite v1.0.1-0.20260329113556-783bbb0820d5 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/libp2p/go-yamux/v4 v4.0.2 // indirect
	github.com/mr-tron/base58 v1.3.0 // indirect
	github.com/xrash/smetrics v0.0.0-20250705151800-55b8f293f342 // indirect
	github.com/zeebo/blake3 v0.2.4 // indirect
	golang.org/x/mod v0.35.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/tools v0.44.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	mvdan.cc/gofumpt v0.9.2 // indirect
)
