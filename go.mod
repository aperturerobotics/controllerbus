module github.com/aperturerobotics/controllerbus

go 1.27.0

require (
	github.com/aperturerobotics/cli v1.1.0 // latest
	github.com/aperturerobotics/common v0.38.4 // latest
	github.com/aperturerobotics/fastjson v0.1.2-0.20261024220211-b654d1d5b5aa
	github.com/aperturerobotics/fsnotify v1.9.1-0.20261120090407-765f22ccaa33
	github.com/aperturerobotics/json-iterator-lite v1.1.0 // master
	github.com/aperturerobotics/protobuf-go-lite v0.19.1-0.20261120061248-f1c7d6295149 // latest
	github.com/aperturerobotics/starpc v0.52.3 // latest
	github.com/aperturerobotics/util v1.34.10-0.20261127192238-8c87e8933f53 // latest
	github.com/mr-tron/base58 v1.3.1
	github.com/pkg/errors v0.9.1
	github.com/sergi/go-diff v1.4.5
	github.com/sirupsen/logrus v1.10.2
	github.com/zeebo/blake3 v0.2.4
	golang.org/x/mod v0.43.0
	golang.org/x/tools v0.53.0
	mvdan.cc/gofumpt v0.12.0
)

require (
	github.com/aperturerobotics/abseil-cpp v0.0.0-20260131110040-4bb56e2f9017 // indirect
	github.com/aperturerobotics/go-protoc-gen-prost v0.0.0-20261024221648-a1c7ae9a2bb8 // indirect
	github.com/aperturerobotics/go-protoc-wasi v0.0.0-20261024220153-869f5fcfca4f // indirect
	github.com/aperturerobotics/go-websocket v1.8.15-0.20260619192713-a096778f08c1 // indirect
	github.com/aperturerobotics/protobuf v0.0.0-20260203024654-8201686529c4 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/libp2p/go-yamux/v5 v5.1.0 // indirect
	github.com/stretchr/testify v1.12.1 // indirect
	github.com/tetratelabs/wazero v1.12.0 // indirect
	github.com/xrash/smetrics v0.0.0-20250705151800-55b8f293f342 // indirect
	golang.org/x/sync v0.25.0 // indirect
	golang.org/x/sys v0.50.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/sirupsen/logrus => github.com/aperturerobotics/logrus v1.9.5-0.20260430110313-9c892333814d
