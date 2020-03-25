module github.com/keys-pub/keysd/service

go 1.12

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.3.5 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/keybase/go-keychain v0.0.0-20200218013740-86d4642e4ce2 // indirect
	github.com/keybase/saltpack v0.0.0-20200228190633-d75baa96bffb // indirect
	github.com/keys-pub/keys v0.0.0-20200324163809-892a74504eee
	github.com/keys-pub/keysd/db v0.0.0-20200321220344-23c2678e5f14
	github.com/keys-pub/keysd/http/api v0.0.0-20200324205758-903123ffbef9
	github.com/keys-pub/keysd/http/client v0.0.0-20200325021855-4d1c8d6feb0f
	github.com/keys-pub/keysd/http/server v0.0.0-20200325202038-b8ae278ae932
	github.com/keys-pub/keysd/wormhole v0.0.0-20200325202038-b8ae278ae932
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/mercari/go-grpc-interceptor v0.0.0-20180110035004-b8ad3827e82a
	github.com/mitchellh/go-ps v1.0.0
	github.com/onsi/ginkgo v1.9.0 // indirect
	github.com/onsi/gomega v1.7.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.5.1
	github.com/urfave/cli v1.22.3
	golang.org/x/crypto v0.0.0-20200320181102-891825fb96df
	golang.org/x/net v0.0.0-20200320220750-118fecf932d8
	golang.org/x/sys v0.0.0-20200321134203-328b4cd54aae // indirect
	google.golang.org/genproto v0.0.0-20200319113533-08878b785e9c // indirect
	google.golang.org/grpc v1.28.0
)

// replace github.com/keys-pub/keys => ../../keys

// replace github.com/keys-pub/keysd/db => ../db

// replace github.com/keys-pub/keysd/http/api => ../http/api

// replace github.com/keys-pub/keysd/http/client => ../http/client

// replace github.com/keys-pub/keysd/http/server => ../http/server

// replace github.com/keys-pub/keysd/wormhole => ../wormhole
