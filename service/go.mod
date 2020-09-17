module github.com/keys-pub/keys-ext/service

go 1.14

require (
	github.com/alta/protopatch v0.0.0-20200702232458-c2bd0c612764
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/keys-pub/keys v0.1.6-0.20200917203958-7a0a1f05c988
	github.com/keys-pub/keys-ext/auth/fido2 v0.0.0-20200913225123-a7320bbf5a79
	github.com/keys-pub/keys-ext/auth/mock v0.0.0-20200909210040-fd6d944acf60
	github.com/keys-pub/keys-ext/http/api v0.0.0-20200917181034-847212d96be0
	github.com/keys-pub/keys-ext/http/client v0.0.0-20200917183732-707b83c90b21
	github.com/keys-pub/keys-ext/http/server v0.0.0-20200917181212-27ed87b9c3b2
	github.com/keys-pub/keys-ext/sdb v0.0.0-20200911204159-4d4aeeb20971
	github.com/keys-pub/keys-ext/vault v0.0.0-20200917183732-707b83c90b21
	github.com/keys-pub/keys-ext/wormhole v0.0.0-20200911204159-4d4aeeb20971
	github.com/mercari/go-grpc-interceptor v0.0.0-20180110035004-b8ad3827e82a
	github.com/mitchellh/go-ps v1.0.0
	github.com/pion/sctp v1.7.10 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	github.com/urfave/cli v1.22.4
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a
	google.golang.org/grpc v1.32.0
	google.golang.org/protobuf v1.25.0
)

// replace github.com/keys-pub/keys => ../../keys

// replace github.com/keys-pub/keys-ext/sdb => ../sdb

// replace github.com/keys-pub/keys-ext/auth/fido2 => ../auth/fido2

// replace github.com/keys-pub/keys-ext/auth/mock => ../auth/mock

// replace github.com/keys-pub/keys-ext/http/api => ../http/api

// replace github.com/keys-pub/keys-ext/http/client => ../http/client

// replace github.com/keys-pub/keys-ext/http/server => ../http/server

// replace github.com/keys-pub/keys-ext/wormhole => ../wormhole

// replace github.com/keys-pub/keys-ext/vault => ../vault
