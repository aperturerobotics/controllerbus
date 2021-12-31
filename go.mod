module github.com/aperturerobotics/controllerbus

go 1.16

// aperture: use protobuf 1.3.x based fork for compatibility
replace (
	github.com/golang/protobuf => github.com/aperturerobotics/go-protobuf-1.3.x v0.0.0-20200726220404-fa7f51c52df0 // aperture-1.3.x
	github.com/sirupsen/logrus => github.com/paralin/logrus v1.8.2-0.20211213222417-cf58d81385c7 // aperture
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55
	google.golang.org/grpc => google.golang.org/grpc v1.30.0
)

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/cenkalti/backoff v1.1.1-0.20190506075156-2146c9339422
	github.com/fsnotify/fsnotify v1.5.0
	github.com/ghodss/yaml v1.0.0
	github.com/golang/protobuf v1.4.2
	github.com/minio/highwayhash v1.0.2
	github.com/mr-tron/base58 v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/sergi/go-diff v1.1.0
	github.com/sirupsen/logrus v1.8.0
	github.com/stretchr/testify v1.7.0 // indirect
	github.com/urfave/cli v1.22.5
	golang.org/x/mod v0.5.1
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.1.8
	google.golang.org/grpc v1.30.0
	mvdan.cc/gofumpt v0.2.1
)
