module github.com/keys-pub/keysd

go 1.12

require (
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/keys-pub/keysd/service v0.0.0-20200220194145-a8f03c5a1979
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	golang.org/x/sys v0.0.0-20200122134326-e047566fdf82 // indirect
	google.golang.org/genproto v0.0.0-20191230161307-f3c370f40bfb // indirect
	google.golang.org/grpc v1.27.0 // indirect
)

replace github.com/keys-pub/keysd/service => ./service

// replace github.com/keys-pub/keysd/http/api => ./http/api

// replace github.com/keys-pub/keysd/http/client => ./http/client

// replace github.com/keys-pub/keysd/http/server => ./http/server

// replace github.com/keys-pub/keysd/db => ./db

// replace github.com/keys-pub/keys => ../keys
