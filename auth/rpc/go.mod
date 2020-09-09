module github.com/keys-pub/keys-ext/auth/rpc

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/uuid v1.1.1
	github.com/keys-pub/go-libfido2 v1.4.1-0.20200603002038-2d73e4e1e232
	github.com/keys-pub/keys-ext/auth/fido2 v0.0.0-20200618211325-4c2d562cade7
	github.com/kr/pretty v0.1.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	google.golang.org/grpc v1.29.1
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

replace github.com/keys-pub/keys-ext/auth/fido2 => ../fido2
