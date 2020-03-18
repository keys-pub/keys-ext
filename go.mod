module github.com/keys-pub/keysd

go 1.12

require (
	github.com/golang/protobuf v1.3.5 // indirect
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/googleapis/gax-go v2.0.2+incompatible // indirect
	github.com/keys-pub/keysd/db v0.0.0-20200318174216-c4166f689f5e // indirect
	github.com/keys-pub/keysd/http/api v0.0.0-20200318174216-c4166f689f5e // indirect
	github.com/keys-pub/keysd/http/client v0.0.0-20200318174216-c4166f689f5e // indirect
	github.com/keys-pub/keysd/service v0.0.0-20200318172516-3411728d0aaa
	github.com/keys-pub/keysd/wormhole v0.0.0-20200318172516-3411728d0aaa // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.5.1
	github.com/urfave/cli v1.22.3 // indirect
	google.golang.org/api v0.20.0 // indirect
	google.golang.org/genproto v0.0.0-20200318110522-7735f76e9fa5 // indirect
	google.golang.org/grpc v1.28.0 // indirect
)

replace github.com/keys-pub/keysd/service => ./service

// replace github.com/keys-pub/keysd/http/api => ./http/api

// replace github.com/keys-pub/keysd/http/client => ./http/client

// replace github.com/keys-pub/keysd/http/server => ./http/server

// replace github.com/keys-pub/keysd/db => ./db

// replace github.com/keys-pub/keys => ../keys
