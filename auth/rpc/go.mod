module github.com/keys-pub/keys-ext/auth/rpc

go 1.14

require (
	github.com/alta/protopatch v0.0.0-20201129223125-3bceb77d56ba // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/uuid v1.1.2
	github.com/keys-pub/go-libfido2 v1.5.2-0.20201217024008-6a7caefe31a1
	github.com/keys-pub/keys-ext/auth/fido2 v0.0.0-20201218004539-7f8afbf3d488
	github.com/kr/pretty v0.1.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	golang.org/x/net v0.0.0-20201216054612-986b41b23924 // indirect
	golang.org/x/sys v0.0.0-20201218084310-7d0127a74742 // indirect
	golang.org/x/text v0.3.4 // indirect
	google.golang.org/genproto v0.0.0-20201214200347-8c77b98c765d // indirect
	google.golang.org/grpc v1.34.0
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

// replace github.com/keys-pub/keys-ext/auth/fido2 => ../fido2

// replace github.com/keys-pub/go-libfido2 => ../../../go-libfido2
