module github.com/aperturerobotics/controllerbus

go 1.22

// This fork uses protobuf-go-lite. This replace can be safely removed but optimizes binary size.
replace github.com/libp2p/go-libp2p => github.com/aperturerobotics/go-libp2p v0.33.1-0.20240504075939-591fc65373be // aperture

require (
	github.com/aperturerobotics/common v0.15.5 // latest
	github.com/aperturerobotics/json-iterator-lite v1.0.0 // latest
	github.com/aperturerobotics/protobuf-go-lite v0.6.1 // latest
	github.com/aperturerobotics/starpc v0.32.4 // main
	github.com/aperturerobotics/util v1.23.0 // latest
)

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/fsnotify/fsnotify v1.7.0
	github.com/ghodss/yaml v1.0.0
	github.com/mr-tron/base58 v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/sergi/go-diff v1.3.1
	github.com/sirupsen/logrus v1.9.3
	github.com/urfave/cli/v2 v2.27.2
	github.com/zeebo/blake3 v0.2.3
	golang.org/x/exp v0.0.0-20240416160154-fe59bbe5cc7f
	golang.org/x/mod v0.17.0
	golang.org/x/tools v0.20.0
	mvdan.cc/gofumpt v0.6.0
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/ipfs/go-cid v0.4.1 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/libp2p/go-libp2p v0.33.2 // indirect
	github.com/libp2p/go-yamux/v4 v4.0.2-0.20240322071716-53ef5820bd48 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.2.0 // indirect
	github.com/multiformats/go-multiaddr v0.12.3 // indirect
	github.com/multiformats/go-multibase v0.2.0 // indirect
	github.com/multiformats/go-multicodec v0.9.0 // indirect
	github.com/multiformats/go-multihash v0.2.3 // indirect
	github.com/multiformats/go-multistream v0.5.0 // indirect
	github.com/multiformats/go-varint v0.0.7 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/spaolacci/murmur3 v1.1.1-0.20190317074736-539464a789e9 // indirect
	github.com/xrash/smetrics v0.0.0-20240312152122-5f08fbb34913 // indirect
	golang.org/x/crypto v0.19.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	lukechampine.com/blake3 v1.2.1 // indirect
	nhooyr.io/websocket v1.8.11 // indirect
)
