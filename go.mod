module github.com/aperturerobotics/controllerbus

go 1.13

// aperture: use protobuf 1.3.x based fork for compatibility
replace (
	github.com/golang/protobuf => github.com/aperturerobotics/go-protobuf-1.3.x v0.0.0-20200706003739-05fb54d407a9 // aperture-1.3.x
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/fsnotify/fsnotify v1.4.9
	github.com/ghodss/yaml v1.0.0
	github.com/golang/protobuf v1.4.2
	github.com/minio/highwayhash v1.0.0
	github.com/mr-tron/base58 v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/sergi/go-diff v1.1.0
	github.com/sirupsen/logrus v1.6.0
	github.com/urfave/cli v1.22.4
	golang.org/x/mod v0.3.0
	golang.org/x/tools v0.0.0-20200626171337-aa94e735be7f
	google.golang.org/grpc v1.29.1
	mvdan.cc/gofumpt v0.0.0-20200627213337-90206bd98491
)
