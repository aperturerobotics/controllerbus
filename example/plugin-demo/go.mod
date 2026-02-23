module github.com/aperturerobotics/controllerbus/example/plugin-demo

go 1.25

toolchain go1.26.0

replace github.com/aperturerobotics/controllerbus => ../..

require (
	github.com/aperturerobotics/controllerbus v0.52.3 // master
	github.com/aperturerobotics/protobuf-go-lite v0.12.2 // latest
	github.com/aperturerobotics/starpc v0.46.2 // indirect; main
	github.com/aperturerobotics/util v1.32.4 // indirect; latest
)

require (
	github.com/blang/semver/v4 v4.0.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.4
)

require (
	github.com/aperturerobotics/cli v1.1.0 // indirect
	github.com/aperturerobotics/json-iterator-lite v1.0.1-0.20260219012250-5319a9708f4a // indirect
	github.com/coder/websocket v1.8.14 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/ipfs/go-cid v0.5.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/libp2p/go-libp2p v0.47.0 // indirect
	github.com/libp2p/go-yamux/v4 v4.0.2 // indirect
	github.com/libp2p/go-yamux/v5 v5.0.1 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.2.0 // indirect
	github.com/multiformats/go-multiaddr v0.16.0 // indirect
	github.com/multiformats/go-multibase v0.2.0 // indirect
	github.com/multiformats/go-multicodec v0.9.1 // indirect
	github.com/multiformats/go-multihash v0.2.3 // indirect
	github.com/multiformats/go-multistream v0.6.1 // indirect
	github.com/multiformats/go-varint v0.0.7 // indirect
	github.com/spaolacci/murmur3 v1.1.1-0.20190317074736-539464a789e9 // indirect
	github.com/xrash/smetrics v0.0.0-20250705151800-55b8f293f342 // indirect
	github.com/zeebo/blake3 v0.2.4 // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/exp v0.0.0-20250606033433-dcc06ee1d476 // indirect
	golang.org/x/mod v0.33.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/tools v0.42.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	lukechampine.com/blake3 v1.4.1 // indirect
	mvdan.cc/gofumpt v0.9.2 // indirect
)
