module github.com/aperturerobotics/controllerbus/example/plugin-demo

go 1.24.0

toolchain go1.25.3

replace github.com/aperturerobotics/controllerbus => ../..

// While not absolutely necessary, this fork uses go-protobuf-lite.
replace github.com/libp2p/go-libp2p => github.com/aperturerobotics/go-libp2p v0.37.1-0.20241111002741-5cfbb50b74e0 // aperture

require (
	github.com/aperturerobotics/controllerbus v0.51.0 // master
	github.com/aperturerobotics/protobuf-go-lite v0.11.0 // latest
	github.com/aperturerobotics/starpc v0.39.9 // indirect; main
	github.com/aperturerobotics/util v1.31.3 // indirect; latest
)

require (
	github.com/blang/semver/v4 v4.0.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/aperturerobotics/cli v1.0.0 // indirect
	github.com/aperturerobotics/common v0.22.12 // indirect
	github.com/aperturerobotics/json-iterator-lite v1.0.1-0.20250712004945-4e5f8882b0b8 // indirect
	github.com/coder/websocket v1.8.14 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/ipfs/go-cid v0.4.1 // indirect
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/libp2p/go-libp2p v0.43.0 // indirect
	github.com/libp2p/go-yamux/v4 v4.0.2-0.20240826150533-e92055b23e0e // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.2.0 // indirect
	github.com/multiformats/go-multiaddr v0.13.0 // indirect
	github.com/multiformats/go-multibase v0.2.0 // indirect
	github.com/multiformats/go-multihash v0.2.3 // indirect
	github.com/multiformats/go-multistream v0.5.0 // indirect
	github.com/multiformats/go-varint v0.0.7 // indirect
	github.com/spaolacci/murmur3 v1.1.1-0.20190317074736-539464a789e9 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	github.com/zeebo/blake3 v0.2.4 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/exp v0.0.0-20250408133849-7e4ce0ab07d0 // indirect
	golang.org/x/mod v0.28.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/tools v0.37.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	lukechampine.com/blake3 v1.3.0 // indirect
	mvdan.cc/gofumpt v0.9.1 // indirect
)
