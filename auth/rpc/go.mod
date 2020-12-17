module github.com/keys-pub/keys-ext/auth/rpc

go 1.14

require (
	github.com/alta/protopatch v0.0.0-20201129223125-3bceb77d56ba // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/uuid v1.1.2
	github.com/keys-pub/go-libfido2 v1.5.2-0.20201217024008-6a7caefe31a1
	github.com/keys-pub/keys-ext/auth/fido2 v0.0.0-20201209192109-53aa0abbc2f3
	github.com/kr/pretty v0.1.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	golang.org/x/net v0.0.0-20201209123823-ac852fbbde11 // indirect
	golang.org/x/sys v0.0.0-20201207223542-d4d67f95c62d // indirect
	golang.org/x/text v0.3.4 // indirect
	google.golang.org/genproto v0.0.0-20201209185603-f92720507ed4 // indirect
	google.golang.org/grpc v1.34.0
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

// replace github.com/keys-pub/keys-ext/auth/fido2 => ../fido2

// replace github.com/keys-pub/go-libfido2 => ../../../go-libfido2
