module github.com/keys-pub/keys-ext/service

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/keys-pub/keys v0.1.2-0.20200714015424-b54f7f572bd1
	github.com/keys-pub/keys-ext/auth/fido2 v0.0.0-20200618211325-4c2d562cade7
	github.com/keys-pub/keys-ext/auth/mock v0.0.0-20200618212723-bf12ba4cbdc4
	github.com/keys-pub/keys-ext/http/api v0.0.0-20200704211016-ce8ce10a1087
	github.com/keys-pub/keys-ext/http/client v0.0.0-20200716194119-b796152d6e47
	github.com/keys-pub/keys-ext/http/server v0.0.0-20200714015603-7e7d65871956
	github.com/keys-pub/keys-ext/sdb v0.0.0-20200704211703-900392aae3e8
	github.com/keys-pub/keys-ext/vault v0.0.0-20200716214240-2f8ab4ed91de
	github.com/keys-pub/keys-ext/wormhole v0.0.0-20200624011543-a01a0028982e
	github.com/mercari/go-grpc-interceptor v0.0.0-20180110035004-b8ad3827e82a
	github.com/mitchellh/go-ps v1.0.0
	github.com/olekukonko/tablewriter v0.0.4 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	github.com/urfave/cli v1.22.4
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899
	google.golang.org/grpc v1.29.1
	google.golang.org/protobuf v1.24.0 // indirect
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
