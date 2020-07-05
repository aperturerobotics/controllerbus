module compile-module

go 1.14

replace github.com/aperturerobotics/controllerbus => ../..

require (
	github.com/aperturerobotics/controllerbus v0.0.0-00010101000000-000000000000
	github.com/blang/semver v3.5.1+incompatible
	github.com/golang/protobuf v1.4.2
	github.com/minio/highwayhash v1.0.0 // indirect
	github.com/sirupsen/logrus v1.6.0
	golang.org/x/tools v0.0.0-20200626171337-aa94e735be7f
)
