module github.com/keys-pub/keysd

go 1.12

require (
	github.com/keys-pub/keysd/db v0.0.0-20191206000922-ce6885426ccc // indirect
	github.com/keys-pub/keysd/http/api v0.0.0-20191206000922-ce6885426ccc // indirect
	github.com/keys-pub/keysd/http/client v0.0.0-20191206000922-ce6885426ccc // indirect
	github.com/keys-pub/keysd/service v0.0.0-20191206000922-ce6885426ccc
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
)

// replace github.com/keys-pub/keysd/service => ./service

// replace github.com/keys-pub/keysd/http/api => ./http/api

// replace github.com/keys-pub/keysd/http/client => ./http/client

// replace github.com/keys-pub/keysd/http/server => ./http/server

// replace github.com/keys-pub/keysd/db => ./db

// replace github.com/keys-pub/keys => ../keys
