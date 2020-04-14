module github.com/keys-pub/keysd/service

go 1.12

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.3.5 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/keys-pub/keys v0.0.0-20200414155340-f3f1d54a3cd3
	github.com/keys-pub/keysd/db v0.0.0-20200413003215-f85e85366c95
	github.com/keys-pub/keysd/http/api v0.0.0-20200412190331-0e28c0a8f66f
	github.com/keys-pub/keysd/http/client v0.0.0-20200413004457-9dcc3d959be4
	github.com/keys-pub/keysd/http/server v0.0.0-20200413003542-7dbbe8346758
	github.com/keys-pub/keysd/wormhole v0.0.0-20200413004603-509fcacd8791
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/mercari/go-grpc-interceptor v0.0.0-20180110035004-b8ad3827e82a
	github.com/mitchellh/go-ps v1.0.0
	github.com/onsi/gomega v1.7.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.5.1
	github.com/urfave/cli v1.22.3
	golang.org/x/crypto v0.0.0-20200406173513-056763e48d71
	google.golang.org/genproto v0.0.0-20200319113533-08878b785e9c // indirect
	google.golang.org/grpc v1.28.0
)

// replace github.com/keys-pub/keys => ../../keys

// replace github.com/keys-pub/keysd/db => ../db

// replace github.com/keys-pub/keysd/http/api => ../http/api

// replace github.com/keys-pub/keysd/http/client => ../http/client

// replace github.com/keys-pub/keysd/http/server => ../http/server

// replace github.com/keys-pub/keysd/wormhole => ../wormhole
