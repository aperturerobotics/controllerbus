module github.com/aperturerobotics/controllerbus

go 1.27.0

require (
	github.com/aperturerobotics/cli v1.1.0 // latest
	github.com/aperturerobotics/common v0.35.4 // latest
	github.com/aperturerobotics/fastjson v0.1.1
	github.com/aperturerobotics/fsnotify v1.9.1-0.20260506231828-931cb4bf1761
	github.com/aperturerobotics/json-iterator-lite v1.1.0 // master
	github.com/aperturerobotics/protobuf-go-lite v0.18.0 // latest
	github.com/aperturerobotics/starpc v0.52.1 // latest
	github.com/aperturerobotics/util v1.34.10-0.20260802062101-496aab6cefd2 // latest
	github.com/mr-tron/base58 v1.3.0
	github.com/pkg/errors v0.9.1
	github.com/sergi/go-diff v1.4.0
	github.com/sirupsen/logrus v1.10.2
	github.com/zeebo/blake3 v0.2.4
	golang.org/x/mod v0.41.0
	golang.org/x/tools v0.50.0
	mvdan.cc/gofumpt v0.12.0
)

require (
	github.com/aperturerobotics/abseil-cpp v0.0.0-20260131110040-4bb56e2f9017 // indirect
	github.com/aperturerobotics/go-websocket v1.8.15-0.20260619192713-a096778f08c1 // indirect
	github.com/aperturerobotics/protobuf v0.0.0-20260203024654-8201686529c4 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/libp2p/go-yamux/v5 v5.1.0 // indirect
	github.com/stretchr/testify v1.12.1 // indirect
	github.com/xrash/smetrics v0.0.0-20250705151800-55b8f293f342 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
)

replace github.com/sirupsen/logrus => github.com/aperturerobotics/logrus v1.9.5-0.20260430110313-9c892333814d
