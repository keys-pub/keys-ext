module github.com/keys-pub/keys-ext/service

go 1.14

require (
	github.com/alta/protopatch v0.0.0-20201016184603-76d4a1d79afd
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.4.3
	github.com/golang/snappy v0.0.2 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/keys-pub/keys v0.1.18-0.20201117233052-2bfb5f4d6161
	github.com/keys-pub/keys-ext/auth/fido2 v0.0.0-20201118000009-28224424e74f
	github.com/keys-pub/keys-ext/auth/mock v0.0.0-20201018000238-7b6186f1fe97
	github.com/keys-pub/keys-ext/http/api v0.0.0-20201117233622-e7ee764fc003
	github.com/keys-pub/keys-ext/http/client v0.0.0-20201117235545-32432664de26
	github.com/keys-pub/keys-ext/http/server v0.0.0-20201117233622-e7ee764fc003
	github.com/keys-pub/keys-ext/sdb v0.0.0-20201118000009-28224424e74f
	github.com/keys-pub/keys-ext/vault v0.0.0-20201118000009-28224424e74f
	github.com/keys-pub/keys-ext/wormhole v0.0.0-20201118000009-28224424e74f
	github.com/keys-pub/keys-ext/ws/api v0.0.0-20201118000009-28224424e74f // indirect
	github.com/keys-pub/keys-ext/ws/client v0.0.0-20201118000009-28224424e74f
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/mercari/go-grpc-interceptor v0.0.0-20180110035004-b8ad3827e82a
	github.com/mitchellh/go-ps v1.0.0
	github.com/pion/sctp v1.7.11 // indirect
	github.com/pkg/errors v0.9.1
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/urfave/cli v1.22.5
	github.com/vmihailenco/msgpack/v4 v4.3.12
	github.com/vmihailenco/tagparser v0.1.2 // indirect
	golang.org/x/crypto v0.0.0-20201117144127-c1f2f97bffc9
	google.golang.org/appengine v1.6.7
	google.golang.org/genproto v0.0.0-20201117123952-62d171c70ae1 // indirect
	google.golang.org/grpc v1.33.2
	google.golang.org/protobuf v1.25.0
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

// replace github.com/keys-pub/keys-ext/ws => ../ws
