module github.com/aperturerobotics/controllerbus

go 1.19

replace (
	github.com/sirupsen/logrus => github.com/aperturerobotics/logrus v1.9.1-0.20221224130652-ff61cbb763af // aperture
	google.golang.org/protobuf => github.com/aperturerobotics/protobuf-go v1.30.1-0.20230428014030-7089409cbc63 // aperture
)

require (
	github.com/aperturerobotics/starpc v0.19.2 // latest
	github.com/aperturerobotics/util v1.5.4 // latest
	github.com/blang/semver v3.5.1+incompatible
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/fsnotify/fsnotify v1.6.0
	github.com/ghodss/yaml v1.0.0
	github.com/mr-tron/base58 v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/sergi/go-diff v1.3.1
	github.com/sirupsen/logrus v1.9.3
	github.com/urfave/cli/v2 v2.25.7
	github.com/valyala/fastjson v1.6.4
	github.com/zeebo/blake3 v0.2.3
	golang.org/x/exp v0.0.0-20230711153332-06a737ee72cb
	golang.org/x/mod v0.12.0
	golang.org/x/tools v0.11.0
	google.golang.org/protobuf v1.31.0
	mvdan.cc/gofumpt v0.5.0
)

require github.com/Jeffail/gabs/v2 v2.7.0

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/ipfs/go-cid v0.4.1 // indirect
	github.com/klauspost/compress v1.16.5 // indirect
	github.com/klauspost/cpuid/v2 v2.2.5 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/libp2p/go-libp2p v0.28.0 // indirect
	github.com/libp2p/go-yamux/v4 v4.0.1 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.2.0 // indirect
	github.com/multiformats/go-multiaddr v0.9.0 // indirect
	github.com/multiformats/go-multibase v0.2.0 // indirect
	github.com/multiformats/go-multicodec v0.9.0 // indirect
	github.com/multiformats/go-multihash v0.2.2 // indirect
	github.com/multiformats/go-multistream v0.4.1 // indirect
	github.com/multiformats/go-varint v0.0.7 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/spaolacci/murmur3 v1.1.1-0.20190317074736-539464a789e9 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	golang.org/x/crypto v0.7.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	lukechampine.com/blake3 v1.2.1 // indirect
	nhooyr.io/websocket v1.8.8-0.20221213223501-14fb98eba64e // indirect
)
