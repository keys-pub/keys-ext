module github.com/keys-pub/keys-ext/service

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/keys-pub/keys v0.0.0-20200618211112-96955ab2a908
	github.com/keys-pub/keys-ext/auth/fido2 v0.0.0-20200618211325-4c2d562cade7
	github.com/keys-pub/keys-ext/auth/mock v0.0.0-20200618212723-bf12ba4cbdc4
	github.com/keys-pub/keys-ext/http/api v0.0.0-20200618211624-e8000cad93a4
	github.com/keys-pub/keys-ext/http/client v0.0.0-20200618212627-7af53dede812
	github.com/keys-pub/keys-ext/http/server v0.0.0-20200618211917-c3daf3e8ad10
	github.com/keys-pub/keys-ext/sdb v0.0.0-20200618212405-e2ddc9614ef5
	github.com/keys-pub/keys-ext/vault v0.0.0-20200618212627-7af53dede812
	github.com/keys-pub/keys-ext/wormhole v0.0.0-20200618212723-bf12ba4cbdc4
	github.com/mercari/go-grpc-interceptor v0.0.0-20180110035004-b8ad3827e82a
	github.com/mitchellh/go-ps v1.0.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	github.com/urfave/cli v1.22.4
	golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9
	google.golang.org/grpc v1.29.1
)

// replace github.com/keys-pub/keys => ../../keys

// replace github.com/keys-pub/keys-ext/sdb => ../sdb

// replace github.com/keys-pub/keys-ext/auth/fido2 => ../auth/fido2

// replace github.com/keys-pub/keys-ext/auth/mock => ../auth/mock

// replace github.com/keys-pub/keys-ext/http/api => ../http/api

// replace github.com/keys-pub/keys-ext/http/client => ../http/client

// replace github.com/keys-pub/keys-ext/http/server => ../http/server

// replace github.com/keys-pub/keys-ext/wormhole => ../wormhole

// replace github.com/keys-pub/keys-ext/vault => ../vault
