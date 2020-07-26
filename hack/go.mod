module github.com/aperturerobotics/genproto/tools

go 1.13

replace github.com/multiformats/go-multihash => github.com/paralin/go-multihash v0.0.11-0.20200526102400-a989a5c6678b // gopherjs-compat

// aperture: use protobuf 1.3.x based fork for compatibility
replace (
	github.com/golang/protobuf => github.com/aperturerobotics/go-protobuf-1.3.x v0.0.0-20200706003739-05fb54d407a9 // aperture-1.3.x
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)

require (
	github.com/aperturerobotics/controllerbus v0.6.1
	github.com/blang/semver v3.5.1+incompatible
	github.com/golang/protobuf v1.4.2
	github.com/kr/text v0.2.0 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/square/goprotowrap v0.0.0-20190116012208-bb93590db2db
	github.com/stretchr/testify v1.6.1 // indirect
	golang.org/x/sys v0.0.0-20200323222414-85ca7c5b95cd // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
)
