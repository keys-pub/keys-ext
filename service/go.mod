module github.com/keys-pub/keysd/service

go 1.12

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.3.4 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/keys-pub/keys v0.0.0-20200318230220-d1e9a6a02f2f
	github.com/keys-pub/keysd/db v0.0.0-20200306174951-faa8a8074ae0
	github.com/keys-pub/keysd/http/api v0.0.0-20200317224602-68134b1264db
	github.com/keys-pub/keysd/http/client v0.0.0-20200317224602-68134b1264db
	github.com/keys-pub/keysd/http/server v0.0.0-20200317222721-717bf70f4f22
	github.com/keys-pub/keysd/wormhole v0.0.0-20200318175959-d85e63958206
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/mercari/go-grpc-interceptor v0.0.0-20180110035004-b8ad3827e82a
	github.com/mitchellh/go-ps v1.0.0
	github.com/onsi/ginkgo v1.9.0 // indirect
	github.com/onsi/gomega v1.7.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.5.1
	github.com/urfave/cli v1.22.2
	golang.org/x/crypto v0.0.0-20200317142112-1b76d66859c6
	google.golang.org/genproto v0.0.0-20200306153348-d950eab6f860 // indirect
	google.golang.org/grpc v1.27.1
)

// replace github.com/keys-pub/keys => ../../keys

// replace github.com/keys-pub/keysd/db => ../db

// replace github.com/keys-pub/keysd/http/api => ../http/api

// replace github.com/keys-pub/keysd/http/client => ../http/client

// replace github.com/keys-pub/keysd/http/server => ../http/server

replace github.com/keys-pub/keysd/wormhole => ../wormhole
