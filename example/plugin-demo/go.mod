module github.com/aperturerobotics/controllerbus/example/plugin-demo

go 1.16

replace github.com/aperturerobotics/controllerbus => ../..

// aperture: use protobuf 1.3.x based fork for compatibility
replace (
	github.com/golang/protobuf => github.com/aperturerobotics/go-protobuf-1.3.x v0.0.0-20200726220404-fa7f51c52df0 // aperture-1.3.x
	github.com/sirupsen/logrus => github.com/paralin/logrus v1.8.2-0.20211213222417-cf58d81385c7 // aperture
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55
	google.golang.org/grpc => google.golang.org/grpc v1.30.0
)

require (
	github.com/aperturerobotics/controllerbus v0.0.0-00010101000000-000000000000
	github.com/blang/semver v3.5.1+incompatible
	github.com/golang/protobuf v1.4.2
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.0
)
