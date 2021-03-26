module github.com/keys-pub/keys-ext/auth/rpc

go 1.14

require (
	github.com/alta/protopatch v0.3.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/uuid v1.2.0
	github.com/keys-pub/go-libfido2 v1.5.2
	github.com/keys-pub/keys-ext/auth/fido2 v0.0.0-20210307192523-fc1e66f7c97e
	github.com/kr/pretty v0.1.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
	golang.org/x/net v0.0.0-20210324205630-d1beb07c2056 // indirect
	google.golang.org/genproto v0.0.0-20210325224202-eed09b1b5210 // indirect
	google.golang.org/grpc v1.36.1
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

// replace github.com/keys-pub/keys-ext/auth/fido2 => ../fido2

// replace github.com/keys-pub/go-libfido2 => ../../../go-libfido2
