module github.com/aperturerobotics/controllerbus/example/plugin-demo

go 1.16

replace github.com/aperturerobotics/controllerbus => ../..

// aperture: use protobuf 1.3.x based fork for compatibility
replace github.com/golang/protobuf => github.com/aperturerobotics/go-protobuf-1.3.x v0.0.0-20200706003739-05fb54d407a9 // aperture-1.3.x

require (
	github.com/aperturerobotics/controllerbus v0.0.0-00010101000000-000000000000
	github.com/blang/semver v3.5.1+incompatible
	github.com/golang/protobuf v1.4.2
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.0
)
