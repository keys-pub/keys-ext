module github.com/keys-pub/keysd/fido2

go 1.14

require (
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.0 // indirect
	github.com/keys-pub/go-libfido2 v0.0.0-20200427035032-3e225c0ecafc
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.5.0
	github.com/urfave/cli v1.22.4
	golang.org/x/net v0.0.0-20200421231249-e086a090c8fd // indirect
	golang.org/x/sys v0.0.0-20200420163511-1957bb5e6d1f // indirect
	golang.org/x/text v0.3.2 // indirect
	google.golang.org/genproto v0.0.0-20200424135956-bca184e23272 // indirect
	google.golang.org/grpc v1.29.1
)

// replace github.com/keys-pub/go-libfido2 => ../../go-libfido2
