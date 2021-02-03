module github.com/keys-pub/keys-ext/service

go 1.14

require (
	github.com/alta/protopatch v0.2.0
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.4.3
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/keybase/go-keychain v0.0.0-20201121013009-976c83ec27a6 // indirect
	github.com/keys-pub/keys v0.1.20-0.20210102234532-ff048354dd2b
	github.com/keys-pub/keys-ext/auth/fido2 v0.0.0-20201221200935-cda646bb0b44
	github.com/keys-pub/keys-ext/auth/mock v0.0.0-20201018000238-7b6186f1fe97
	github.com/keys-pub/keys-ext/http/api v0.0.0-20210108232832-6bce3e9124fe
	github.com/keys-pub/keys-ext/http/client v0.0.0-20210109001315-76134396e9aa
	github.com/keys-pub/keys-ext/http/server v0.0.0-20210103224830-a8720c3ddd9b
	github.com/keys-pub/keys-ext/sdb v0.0.0-20210109001315-76134396e9aa
	github.com/keys-pub/keys-ext/vault v0.0.0-20210114210042-9547899d11f5
	github.com/keys-pub/keys-ext/wormhole v0.0.0-20210102030049-6622ca14f3bc
	github.com/keys-pub/keys-ext/ws/api v0.0.0-20210103224830-a8720c3ddd9b
	github.com/keys-pub/keys-ext/ws/client v0.0.0-20210109001315-76134396e9aa
	github.com/mercari/go-grpc-interceptor v0.0.0-20180110035004-b8ad3827e82a
	github.com/mitchellh/go-ps v1.0.0
	github.com/pion/sctp v1.7.11 // indirect
	github.com/pkg/errors v0.9.1
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.6.1
	github.com/urfave/cli v1.22.5
	golang.org/x/term v0.0.0-20201210144234-2321bbc49cbf
	google.golang.org/grpc v1.34.0
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

// replace github.com/keys-pub/keys-ext/ws/api => ../ws/api

// replace github.com/keys-pub/keys-ext/ws/client => ../ws/client
