module github.com/keys-pub/keysd/service

go 1.12

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/gogo/protobuf v1.3.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/keys-pub/keys v0.0.0-20200423195135-6b5d508024d0
	github.com/keys-pub/keysd/db v0.0.0-20200413003215-f85e85366c95
	github.com/keys-pub/keysd/fido2 v0.0.0-20200427220403-902714bb307e
	github.com/keys-pub/keysd/http/api v0.0.0-20200415010142-cfcd41d36dd1
	github.com/keys-pub/keysd/http/client v0.0.0-20200414170516-1e0e72775735
	github.com/keys-pub/keysd/http/server v0.0.0-20200419021946-2d1a17136bd0
	github.com/keys-pub/keysd/wormhole v0.0.0-20200413004603-509fcacd8791
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/mercari/go-grpc-interceptor v0.0.0-20180110035004-b8ad3827e82a
	github.com/mitchellh/go-ps v1.0.0
	github.com/onsi/gomega v1.7.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.5.0
	github.com/stretchr/testify v1.5.1
	github.com/urfave/cli v1.22.4
	golang.org/x/crypto v0.0.0-20200406173513-056763e48d71
	google.golang.org/grpc v1.29.1
)

// replace github.com/keys-pub/keys => ../../keys

// replace github.com/keys-pub/keysd/db => ../db

replace github.com/keys-pub/keysd/fido2 => ../fido2
// replace github.com/keys-pub/go-libfido2 => ../../go-libfido2

// replace github.com/keys-pub/keysd/http/api => ../http/api

// replace github.com/keys-pub/keysd/http/client => ../http/client

// replace github.com/keys-pub/keysd/http/server => ../http/server

// replace github.com/keys-pub/keysd/wormhole => ../wormhole
