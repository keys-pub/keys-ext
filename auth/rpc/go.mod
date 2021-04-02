module github.com/keys-pub/keys-ext/auth/rpc

go 1.14

require (
	github.com/alta/protopatch v0.3.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/uuid v1.2.0
	github.com/keys-pub/go-libfido2 v1.5.3-0.20210401210751-db20b37a1e88
	github.com/keys-pub/keys-ext/auth/fido2 v0.0.0-20210327130412-59e9fcfcf22c
	github.com/kr/pretty v0.1.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
	golang.org/x/net v0.0.0-20210326220855-61e056675ecf // indirect
	golang.org/x/sys v0.0.0-20210326220804-49726bf1d181 // indirect
	google.golang.org/genproto v0.0.0-20210325224202-eed09b1b5210 // indirect
	google.golang.org/grpc v1.36.1
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

// replace github.com/keys-pub/keys-ext/auth/fido2 => ../fido2

// replace github.com/keys-pub/go-libfido2 => ../../../go-libfido2
