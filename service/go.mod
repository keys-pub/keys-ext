module github.com/keys-pub/keysd/service

go 1.12

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/gogo/protobuf v1.3.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/keys-pub/keys v0.0.0-20200517225925-b4d826f558e6
	github.com/keys-pub/keysd/db v0.0.0-20200511180349-f7a683035492
	github.com/keys-pub/keysd/fido2 v0.0.0-20200429024946-ecdf142d9dad
	github.com/keys-pub/keysd/http/api v0.0.0-20200415010142-cfcd41d36dd1
	github.com/keys-pub/keysd/http/client v0.0.0-20200517230154-d4ab6096765d
	github.com/keys-pub/keysd/http/server v0.0.0-20200517230030-c2db50fa6b8b
	github.com/keys-pub/keysd/wormhole v0.0.0-20200517230256-862f60e2266a
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

// replace github.com/keys-pub/keysd/fido2 => ../fido2
// replace github.com/keys-pub/go-libfido2 => ../../go-libfido2

// replace github.com/keys-pub/keysd/http/api => ../http/api

// replace github.com/keys-pub/keysd/http/client => ../http/client

// replace github.com/keys-pub/keysd/http/server => ../http/server

// replace github.com/keys-pub/keysd/wormhole => ../wormhole
