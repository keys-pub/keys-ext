module github.com/keys-pub/keys-ext/auth/rpc

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/uuid v1.1.2
	github.com/keys-pub/go-libfido2 v1.5.1
	github.com/keys-pub/keys-ext/auth/fido2 v0.0.0-20201112015717-fabc8d0009ac
	github.com/kr/pretty v0.1.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b // indirect
	golang.org/x/sys v0.0.0-20201112073958-5cba982894dd // indirect
	golang.org/x/text v0.3.4 // indirect
	google.golang.org/genproto v0.0.0-20201112120144-2985b7af83de // indirect
	google.golang.org/grpc v1.33.2
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

// replace github.com/keys-pub/keys-ext/auth/fido2 => ../fido2

// replace github.com/keys-pub/go-libfido2 => ../../../go-libfido2
