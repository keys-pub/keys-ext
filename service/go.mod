module github.com/keys-pub/keysd/service

go 1.14

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/gogo/protobuf v1.3.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/keys-pub/keys v0.0.0-20200528181135-18e38db61305
	github.com/keys-pub/keysd/auth/fido2 v0.0.0-20200527222136-fe3bbef02231
	github.com/keys-pub/keysd/db v0.0.0-20200527183902-ffb35f491a74
	github.com/keys-pub/keysd/git v0.0.0-20200528181638-f95ae2138e27
	github.com/keys-pub/keysd/http/api v0.0.0-20200527183746-ca4c118983c4
	github.com/keys-pub/keysd/http/client v0.0.0-20200527183209-12da56e5cb21
	github.com/keys-pub/keysd/http/server v0.0.0-20200527183746-ca4c118983c4
	github.com/keys-pub/keysd/wormhole v0.0.0-20200527183902-ffb35f491a74
	github.com/mercari/go-grpc-interceptor v0.0.0-20180110035004-b8ad3827e82a
	github.com/mitchellh/go-ps v1.0.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.5.1
	github.com/urfave/cli v1.22.4
	github.com/vmihailenco/msgpack/v4 v4.3.11
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/grpc v1.29.1
	gortc.io/stun v1.22.2 // indirect
)

// replace github.com/keys-pub/keys => ../../keys

// replace github.com/keys-pub/keysd/db => ../db

// replace github.com/keys-pub/keysd/auth/fido2 => ../auth/fido2

// replace github.com/keys-pub/keysd/git => ../git

// replace github.com/keys-pub/keysd/http/api => ../http/api

// replace github.com/keys-pub/keysd/http/client => ../http/client

// replace github.com/keys-pub/keysd/http/server => ../http/server

// replace github.com/keys-pub/keysd/wormhole => ../wormhole
