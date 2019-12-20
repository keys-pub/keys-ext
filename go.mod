module github.com/keys-pub/keysd

go 1.12

require (
	github.com/keys-pub/keysd/service v0.0.0-20191220200550-ef89bca227f3
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
)

replace github.com/keys-pub/keysd/service => ./service

// replace github.com/keys-pub/keysd/http/api => ./http/api

// replace github.com/keys-pub/keysd/http/client => ./http/client

// replace github.com/keys-pub/keysd/http/server => ./http/server

// replace github.com/keys-pub/keysd/db => ./db

// replace github.com/keys-pub/keys => ../keys
