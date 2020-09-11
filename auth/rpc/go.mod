module github.com/keys-pub/keys-ext/auth/rpc

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/uuid v1.1.2
	github.com/keys-pub/go-libfido2 v1.4.1-0.20200909205858-280a1bda5932
	github.com/keys-pub/keys-ext/auth/fido2 v0.0.0-20200909213631-9edc97fa0f2c
	github.com/kr/pretty v0.1.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	golang.org/x/net v0.0.0-20200904194848-62affa334b73 // indirect
	golang.org/x/sys v0.0.0-20200909081042-eff7692f9009 // indirect
	google.golang.org/genproto v0.0.0-20200910191746-8ad3c7ee2cd1 // indirect
	google.golang.org/grpc v1.32.0
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

replace github.com/keys-pub/keys-ext/auth/fido2 => ../fido2

replace github.com/keys-pub/go-libfido2 => ../../../go-libfido2
