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
	github.com/keys-pub/keys v0.1.7-0.20201022162733-140ebabc2075
	github.com/keys-pub/keys-ext/auth/fido2 v0.0.0-20201023200942-48e5f880045e
	github.com/keys-pub/keys-ext/auth/mock v0.0.0-20201018000238-7b6186f1fe97
	github.com/keys-pub/keys-ext/http/api v0.0.0-20201023200942-48e5f880045e
	github.com/keys-pub/keys-ext/http/client v0.0.0-20201023200942-48e5f880045e
	github.com/keys-pub/keys-ext/http/server v0.0.0-20201019223330-fac0ac0e3560
	github.com/keys-pub/keys-ext/sdb v0.0.0-20201023200942-48e5f880045e
	github.com/keys-pub/keys-ext/vault v0.0.0-20201023200942-48e5f880045e
	github.com/keys-pub/keys-ext/wormhole v0.0.0-20201023200942-48e5f880045e
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/mercari/go-grpc-interceptor v0.0.0-20180110035004-b8ad3827e82a
	github.com/mitchellh/go-ps v1.0.0
	github.com/pion/sctp v1.7.11 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/urfave/cli v1.22.4
	github.com/vmihailenco/tagparser v0.1.2 // indirect
	golang.org/x/crypto v0.0.0-20201016220609-9e8e0b390897
	golang.org/x/net v0.0.0-20201022231255-08b38378de70 // indirect
	golang.org/x/sys v0.0.0-20201022201747-fb209a7c41cd // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20201022181438-0ff5f38871d5 // indirect
	google.golang.org/grpc v1.33.1
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
