module github.com/aperturerobotics/controllerbus

go 1.25

// This fork drops ecdsa, drops secp256k1, adds eddilithium2 and eddilithium3
replace github.com/libp2p/go-libp2p => github.com/aperturerobotics/go-libp2p v0.37.1-0.20241111002741-5cfbb50b74e0 // aperture

replace github.com/libp2p/go-msgio => github.com/aperturerobotics/go-libp2p-msgio v0.0.0-20240511033615-1b69178aa5c8 // aperture

require (
	github.com/aperturerobotics/cli v1.1.0 // latest
	github.com/aperturerobotics/common v0.30.6 // latest
	github.com/aperturerobotics/json-iterator-lite v1.0.1-0.20260219012250-5319a9708f4a // latest
	github.com/aperturerobotics/protobuf-go-lite v0.12.2 // latest
	github.com/aperturerobotics/starpc v0.46.2 // latest
	github.com/aperturerobotics/util v1.32.4 // latest
)

require (
	github.com/blang/semver/v4 v4.0.0
	github.com/fsnotify/fsnotify v1.9.0
	github.com/ghodss/yaml v1.0.0
	github.com/mr-tron/base58 v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/sergi/go-diff v1.4.0
	github.com/sirupsen/logrus v1.9.4
	github.com/zeebo/blake3 v0.2.4
	golang.org/x/mod v0.33.0
	golang.org/x/tools v0.42.0
	mvdan.cc/gofumpt v0.9.2
)

require (
	github.com/aperturerobotics/abseil-cpp v0.0.0-20260131110040-4bb56e2f9017 // indirect
	github.com/aperturerobotics/go-protoc-gen-prost v0.0.0-20260204215916-dc1f0fed8cfc // indirect
	github.com/aperturerobotics/go-protoc-wasi v0.0.0-20260131050911-b5f94b044584 // indirect
	github.com/aperturerobotics/protobuf v0.0.0-20260203024654-8201686529c4 // indirect
	github.com/coder/websocket v1.8.14 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/ipfs/go-cid v0.4.1 // indirect
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/libp2p/go-libp2p v0.47.0 // indirect
	github.com/libp2p/go-yamux/v4 v4.0.2 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.2.0 // indirect
	github.com/multiformats/go-multiaddr v0.13.0 // indirect
	github.com/multiformats/go-multibase v0.2.0 // indirect
	github.com/multiformats/go-multihash v0.2.3 // indirect
	github.com/multiformats/go-multistream v0.5.0 // indirect
	github.com/multiformats/go-varint v0.0.7 // indirect
	github.com/spaolacci/murmur3 v1.1.1-0.20190317074736-539464a789e9 // indirect
	github.com/tetratelabs/wazero v1.11.0 // indirect
	github.com/xrash/smetrics v0.0.0-20250705151800-55b8f293f342 // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/exp v0.0.0-20250408133849-7e4ce0ab07d0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	lukechampine.com/blake3 v1.3.0 // indirect
)
