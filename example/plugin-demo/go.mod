module github.com/aperturerobotics/controllerbus/example/plugin-demo

go 1.19

replace github.com/aperturerobotics/controllerbus => ../..

replace (
	github.com/sirupsen/logrus => github.com/aperturerobotics/logrus v1.9.1-0.20221224130652-ff61cbb763af // aperture
	google.golang.org/protobuf => github.com/aperturerobotics/protobuf-go v1.30.1-0.20230428014030-7089409cbc63 // aperture
)

require (
	github.com/aperturerobotics/controllerbus v0.26.2
	github.com/blang/semver v3.5.1+incompatible
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.0
	google.golang.org/protobuf v1.30.0
)

require (
	github.com/Jeffail/gabs/v2 v2.7.0 // indirect
	github.com/aperturerobotics/starpc v0.19.1 // indirect
	github.com/aperturerobotics/util v1.3.0 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.1.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/ipfs/go-cid v0.3.2 // indirect
	github.com/klauspost/compress v1.15.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/libp2p/go-libp2p v0.26.3 // indirect
	github.com/libp2p/go-yamux/v4 v4.0.1-0.20220919134236-1c09f2ab3ec1 // indirect
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.2.0 // indirect
	github.com/multiformats/go-multiaddr v0.8.0 // indirect
	github.com/multiformats/go-multibase v0.1.2-0.20220823162309-7160a7347ed1 // indirect
	github.com/multiformats/go-multicodec v0.7.1-0.20221017174837-a2baec7ca709 // indirect
	github.com/multiformats/go-multihash v0.2.2-0.20221030163302-608669da49b6 // indirect
	github.com/multiformats/go-multistream v0.4.1 // indirect
	github.com/multiformats/go-varint v0.0.7 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/spaolacci/murmur3 v1.1.1-0.20190317074736-539464a789e9 // indirect
	github.com/urfave/cli/v2 v2.25.3 // indirect
	github.com/valyala/fastjson v1.6.4 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	github.com/zeebo/blake3 v0.2.3 // indirect
	golang.org/x/crypto v0.4.0 // indirect
	golang.org/x/exp v0.0.0-20230425010034-47ecfdc1ba53 // indirect
	golang.org/x/mod v0.10.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/tools v0.8.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	lukechampine.com/blake3 v1.1.8-0.20220321170924-7afca5966e5e // indirect
	mvdan.cc/gofumpt v0.5.0 // indirect
	nhooyr.io/websocket v1.8.8-0.20221213223501-14fb98eba64e // indirect
)
