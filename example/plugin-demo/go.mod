module github.com/aperturerobotics/controllerbus/example/plugin-demo

go 1.16

replace github.com/aperturerobotics/controllerbus => ../..

replace (
	github.com/sirupsen/logrus => github.com/aperturerobotics/logrus v1.9.0 // aperture
	google.golang.org/protobuf => github.com/aperturerobotics/protobuf-go 7fb3cdbd9197 // aperture
)

require (
	github.com/aperturerobotics/controllerbus v0.24.0
	github.com/blang/semver v3.5.1+incompatible
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.0
	google.golang.org/protobuf v1.29.0
)
