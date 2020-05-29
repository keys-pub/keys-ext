module github.com/keys-pub/keys-ext/service

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/gogo/protobuf v1.3.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/keys-pub/keys v0.0.0-20200529222241-8c34c3af1194
	github.com/keys-pub/keys-ext/auth/fido2 v0.0.0-20200528184029-7548f2a0a594
	github.com/keys-pub/keys-ext/db v0.0.0-20200528184029-7548f2a0a594
	github.com/keys-pub/keys-ext/git v0.0.0-20200529223633-ff4b9d73dca6
	github.com/keys-pub/keys-ext/http/api v0.0.0-20200528184029-7548f2a0a594
	github.com/keys-pub/keys-ext/http/client v0.0.0-20200528185501-04f091ec8e61
	github.com/keys-pub/keys-ext/http/server v0.0.0-20200528185324-90ced7e635aa
	github.com/keys-pub/keys-ext/wormhole v0.0.0-20200528185636-49b5d7075454
	github.com/mercari/go-grpc-interceptor v0.0.0-20180110035004-b8ad3827e82a
	github.com/mitchellh/go-ps v1.0.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.5.1
	github.com/urfave/cli v1.22.4
	github.com/vmihailenco/msgpack/v4 v4.3.11
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37
	google.golang.org/grpc v1.29.1
)

// replace github.com/keys-pub/keys => ../../keys

// replace github.com/keys-pub/keys-ext/db => ../db

// replace github.com/keys-pub/keys-ext/auth/fido2 => ../auth/fido2

// replace github.com/keys-pub/keys-ext/git => ../git

// replace github.com/keys-pub/keys-ext/http/api => ../http/api

// replace github.com/keys-pub/keys-ext/http/client => ../http/client

// replace github.com/keys-pub/keys-ext/http/server => ../http/server

// replace github.com/keys-pub/keys-ext/wormhole => ../wormhole
