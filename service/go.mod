module github.com/keys-pub/keys-ext/service

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/keys-pub/keys v0.0.0-20200604182647-b8d6b989cc48
	github.com/keys-pub/keys-ext/auth/fido2 v0.0.0-20200528184029-7548f2a0a594
	github.com/keys-pub/keys-ext/db v0.0.0-20200528184029-7548f2a0a594
	github.com/keys-pub/keys-ext/http/api v0.0.0-20200528184029-7548f2a0a594
	github.com/keys-pub/keys-ext/http/client v0.0.0-20200603163512-e65da09d2d8c
	github.com/keys-pub/keys-ext/http/server v0.0.0-20200528185324-90ced7e635aa
	github.com/keys-pub/keys-ext/wormhole v0.0.0-20200604181719-aff69f455544
	github.com/mercari/go-grpc-interceptor v0.0.0-20180110035004-b8ad3827e82a
	github.com/mitchellh/go-ps v1.0.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.5.1
	github.com/urfave/cli v1.22.4
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37
	google.golang.org/grpc v1.29.1
)

// replace github.com/keys-pub/keys => ../../keys

// replace github.com/keys-pub/keys-ext/db => ../db

// replace github.com/keys-pub/keys-ext/auth/fido2 => ../auth/fido2

// replace github.com/keys-pub/keys-ext/http/api => ../http/api

// replace github.com/keys-pub/keys-ext/http/client => ../http/client

// replace github.com/keys-pub/keys-ext/http/server => ../http/server

// replace github.com/keys-pub/keys-ext/wormhole => ../wormhole
