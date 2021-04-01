module github.com/keys-pub/keys-ext/service

go 1.14

require (
	github.com/alta/protopatch v0.3.3
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.5.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/keys-pub/keys v0.1.21-0.20210331163518-474087d0d185
	github.com/keys-pub/keys-ext/auth/fido2 v0.0.0-20210331211138-7be3b751d8ad
	github.com/keys-pub/keys-ext/auth/mock v0.0.0-20210401184359-d3fda856e211
	github.com/keys-pub/keys-ext/http/api v0.0.0-20210331211138-7be3b751d8ad
	github.com/keys-pub/keys-ext/http/client v0.0.0-20210331211138-7be3b751d8ad
	github.com/keys-pub/keys-ext/http/server v0.0.0-20210329181014-5a15d0a4eba1
	github.com/keys-pub/keys-ext/sdb v0.0.0-20210109001315-76134396e9aa
	github.com/keys-pub/keys-ext/vault v0.0.0-20210331211138-7be3b751d8ad
	github.com/keys-pub/keys-ext/wormhole v0.0.0-20210102030049-6622ca14f3bc
	github.com/keys-pub/keys-ext/ws/client v0.0.0-20210109001315-76134396e9aa
	github.com/mercari/go-grpc-interceptor v0.0.0-20180110035004-b8ad3827e82a
	github.com/mitchellh/go-ps v1.0.0
	github.com/pion/sctp v1.7.11 // indirect
	github.com/pkg/errors v0.9.1
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli v1.22.5
	golang.org/x/term v0.0.0-20210317153231-de623e64d2a6
	google.golang.org/grpc v1.36.1
	google.golang.org/protobuf v1.26.0
	gortc.io/stun v1.23.0 // indirect
)

// replace github.com/keys-pub/keys => ../../keys

// replace github.com/keys-pub/keys-ext/sdb => ../sdb

// replace github.com/keys-pub/keys-ext/auth/fido2 => ../auth/fido2

// replace github.com/keys-pub/keys-ext/auth/mock => ../auth/mock

// replace github.com/keys-pub/keys-ext/http/api => ../http/api

// replace github.com/keys-pub/keys-ext/http/client => ../http/client

// replace github.com/keys-pub/keys-ext/http/server => ../http/server

// replace github.com/keys-pub/keys-ext/vault => ../vault

// replace github.com/keys-pub/keys-ext/wormhole => ../wormhole

// replace github.com/keys-pub/keys-ext/ws/api => ../ws/api

// replace github.com/keys-pub/keys-ext/ws/client => ../ws/client
