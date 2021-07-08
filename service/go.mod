module github.com/keys-pub/keys-ext/service

go 1.14

require (
	github.com/alta/protopatch v0.3.4
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/golang/protobuf v1.5.2
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/keybase/saltpack v0.0.0-20210611181147-9dd0a21addc6 // indirect
	github.com/keys-pub/keys v0.1.22-0.20210708223433-a34d3ce96fb2
	github.com/keys-pub/keys-ext/auth/fido2 v0.0.0-20210708231605-d6538034681a
	github.com/keys-pub/keys-ext/auth/mock v0.0.0-20210401184359-d3fda856e211
	github.com/keys-pub/keys-ext/http/api v0.0.0-20210708231605-d6538034681a
	github.com/keys-pub/keys-ext/http/client v0.0.0-20210708231605-d6538034681a
	github.com/keys-pub/keys-ext/http/server v0.0.0-20210525002537-0c132efd0ef7
	github.com/keys-pub/keys-ext/sdb v0.0.0-20210708231605-d6538034681a
	github.com/keys-pub/keys-ext/vault v0.0.0-20210708223757-e4f0569fdc50
	github.com/keys-pub/keys-ext/wormhole v0.0.0-20210708231605-d6538034681a
	github.com/keys-pub/keys-ext/ws/api v0.0.0-20210402011710-71dc6eac40c7 // indirect
	github.com/keys-pub/keys-ext/ws/client v0.0.0-20210402011710-71dc6eac40c7
	github.com/mercari/go-grpc-interceptor v0.0.0-20180110035004-b8ad3827e82a
	github.com/minio/sio v0.3.0 // indirect
	github.com/mitchellh/go-ps v1.0.0
	github.com/pion/sctp v1.7.12 // indirect
	github.com/pkg/errors v0.9.1
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli v1.22.5
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e // indirect
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e // indirect
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b
	google.golang.org/genproto v0.0.0-20210708141623-e76da96a951f // indirect
	google.golang.org/grpc v1.39.0
	google.golang.org/protobuf v1.27.1
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
