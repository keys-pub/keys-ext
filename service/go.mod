module github.com/keys-pub/keysd/service

go 1.12

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.3.5 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/keys-pub/keys v0.0.0-20200404175525-c68f23ea44c3
	github.com/keys-pub/keysd/db v0.0.0-20200402000046-bfff35e9624b
	github.com/keys-pub/keysd/http/api v0.0.0-20200402183018-a85eceb453b1
	github.com/keys-pub/keysd/http/client v0.0.0-20200402191904-495668794c2d
	github.com/keys-pub/keysd/http/server v0.0.0-20200402191721-a864d6b0e313
	github.com/keys-pub/keysd/wormhole v0.0.0-20200404175609-516ecb7ccd6f
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/mercari/go-grpc-interceptor v0.0.0-20180110035004-b8ad3827e82a
	github.com/mitchellh/go-ps v1.0.0
	github.com/onsi/gomega v1.7.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.5.1
	github.com/urfave/cli v1.22.3
	golang.org/x/crypto v0.0.0-20200323165209-0ec3e9974c59
	google.golang.org/genproto v0.0.0-20200319113533-08878b785e9c // indirect
	google.golang.org/grpc v1.28.0
)

// replace github.com/keys-pub/keys => ../../keys

// replace github.com/keys-pub/keysd/db => ../db

// replace github.com/keys-pub/keysd/http/api => ../http/api

// replace github.com/keys-pub/keysd/http/client => ../http/client

// replace github.com/keys-pub/keysd/http/server => ../http/server

// replace github.com/keys-pub/keysd/wormhole => ../wormhole
