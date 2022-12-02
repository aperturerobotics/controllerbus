module github.com/aperturerobotics/controllerbus

go 1.18

replace (
	github.com/sirupsen/logrus => github.com/aperturerobotics/logrus v1.8.2-0.20220322010420-77ab346a2cf8 // aperture
	google.golang.org/protobuf => github.com/aperturerobotics/protobuf-go v1.28.2-0.20221202092004-7e5a6a8cf680 // aperture
)

require (
	github.com/aperturerobotics/starpc v0.15.2-0.20221202005525-702e2fb770bf
	github.com/aperturerobotics/util v0.0.0-20221202091107-34bf4a704df5
	github.com/blang/semver v3.5.1+incompatible
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/fsnotify/fsnotify v1.6.0
	github.com/ghodss/yaml v1.0.0
	github.com/mr-tron/base58 v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/sergi/go-diff v1.2.0
	github.com/sirupsen/logrus v1.9.0
	github.com/urfave/cli/v2 v2.23.5
	github.com/valyala/fastjson v1.6.3
	github.com/zeebo/blake3 v0.2.3
	golang.org/x/exp v0.0.0-20221126150942-6ab00d035af9
	golang.org/x/mod v0.7.0
	golang.org/x/tools v0.3.1-0.20221201230950-47a82463d369
	google.golang.org/protobuf v1.28.1
	mvdan.cc/gofumpt v0.3.1
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.1.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/ipfs/go-cid v0.3.2 // indirect
	github.com/klauspost/compress v1.15.10 // indirect
	github.com/klauspost/cpuid/v2 v2.1.1 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/libp2p/go-libp2p v0.23.3-0.20221109121032-c334288f8fe4 // indirect
	github.com/libp2p/go-openssl v0.1.0 // indirect
	github.com/libp2p/go-yamux/v4 v4.0.1-0.20220919134236-1c09f2ab3ec1 // indirect
	github.com/mattn/go-pointer v0.0.1 // indirect
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.1.1-0.20220823151017-f5af2eed4d9c // indirect
	github.com/multiformats/go-multiaddr v0.7.0 // indirect
	github.com/multiformats/go-multibase v0.1.2-0.20220823162309-7160a7347ed1 // indirect
	github.com/multiformats/go-multicodec v0.7.1-0.20221017174837-a2baec7ca709 // indirect
	github.com/multiformats/go-multihash v0.2.2-0.20221030163302-608669da49b6 // indirect
	github.com/multiformats/go-varint v0.0.7-0.20220823162201-881f9a52d5d2 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/spacemonkeygo/spacelog v0.0.0-20180420211403-2296661a0572 // indirect
	github.com/spaolacci/murmur3 v1.1.1-0.20190317074736-539464a789e9 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	golang.org/x/crypto v0.2.1-0.20221109165004-21d60a152191 // indirect
	golang.org/x/sys v0.2.0 // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
	lukechampine.com/blake3 v1.1.8-0.20220321170924-7afca5966e5e // indirect
	nhooyr.io/websocket v1.8.8-0.20210410000328-8dee580a7f74 // indirect
)
