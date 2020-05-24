module github.com/keys-pub/keysd/auth/rpc

go 1.14

require (
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/google/uuid v1.1.1
	github.com/keys-pub/go-libfido2 v1.4.1-0.20200522215841-590249e0b34c
	github.com/keys-pub/keysd/auth/fido2 v0.0.0-20200524000041-6d7e23f9bca0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	golang.org/x/net v0.0.0-20200520182314-0ba52f642ac2 // indirect
	golang.org/x/sys v0.0.0-20200519105757-fe76b779f299 // indirect
	golang.org/x/text v0.3.2 // indirect
	google.golang.org/genproto v0.0.0-20200521103424-e9a78aa275b7 // indirect
	google.golang.org/grpc v1.29.1
)

// replace github.com/keys-pub/keysd/auth/fido2 => ../fido2
