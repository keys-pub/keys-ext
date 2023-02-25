module github.com/keys-pub/keys-ext/auth/rpc

go 1.14

require (
	github.com/alta/protopatch v0.3.4 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/uuid v1.2.0
	github.com/keys-pub/go-libfido2 v1.5.3-0.20210401210751-db20b37a1e88
	github.com/keys-pub/keys-ext/auth/fido2 v0.0.0-20210708231605-d6538034681a
	github.com/kr/pretty v0.1.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
	golang.org/x/net v0.7.0 // indirect
	google.golang.org/genproto v0.0.0-20210708141623-e76da96a951f // indirect
	google.golang.org/grpc v1.39.0
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

// replace github.com/keys-pub/keys-ext/auth/fido2 => ../fido2

// replace github.com/keys-pub/go-libfido2 => ../../../go-libfido2
