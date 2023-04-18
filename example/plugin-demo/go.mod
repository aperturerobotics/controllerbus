module github.com/aperturerobotics/controllerbus/example/plugin-demo

go 1.16

replace github.com/aperturerobotics/controllerbus => ../..

replace (
	github.com/sirupsen/logrus => github.com/aperturerobotics/logrus v1.8.2-0.20220322010420-77ab346a2cf8 // aperture
	google.golang.org/protobuf => github.com/aperturerobotics/protobuf-go v1.28.2-0.20221202092004-7e5a6a8cf680 // aperture
)

require (
	github.com/aperturerobotics/controllerbus v0.25.3
	github.com/blang/semver v3.5.1+incompatible
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.0
	google.golang.org/protobuf v1.30.0
)
