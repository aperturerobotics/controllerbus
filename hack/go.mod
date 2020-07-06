module github.com/aperturerobotics/genproto/tools

go 1.13
// aperture: use 1.3.x based fork for compatibility
replace github.com/golang/protobuf => github.com/aperturerobotics/go-protobuf-1.3.x v0.0.0-20200705233748-404297258551 // aperture-1.3.x

require (
	github.com/golang/protobuf v1.3.3-0.20190827175835-822fe56949f5
	github.com/golangci/golangci-lint v1.28.0 // indirect
	github.com/square/goprotowrap v0.0.0-20190116012208-bb93590db2db
)
