module github.com/aperturerobotics/controllerbus

go 1.16

// aperture: use protobuf 1.3.x based fork for compatibility
replace (
	github.com/golang/protobuf => github.com/aperturerobotics/go-protobuf-1.3.x v0.0.0-20200706003739-05fb54d407a9 // aperture-1.3.x
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55
	google.golang.org/grpc => google.golang.org/grpc v1.30.0
)

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/cenkalti/backoff v1.1.1-0.20190506075156-2146c9339422
	github.com/fsnotify/fsnotify v1.5.0
	github.com/ghodss/yaml v1.0.0
	github.com/golang/protobuf v1.4.2
	github.com/kr/pretty v0.2.0 // indirect
	github.com/minio/highwayhash v1.0.2
	github.com/mr-tron/base58 v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/sergi/go-diff v1.1.0
	github.com/sirupsen/logrus v1.8.0
	github.com/stretchr/testify v1.7.0 // indirect
	github.com/urfave/cli v1.22.5
	golang.org/x/mod v0.4.2
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.1.7
	google.golang.org/grpc v1.30.0
	mvdan.cc/gofumpt v0.1.1-0.20210908155448-1727efedc3ad
)
