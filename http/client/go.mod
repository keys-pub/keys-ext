module github.com/keys-pub/keysd/http/client

go 1.12

require (
	github.com/keys-pub/keys v0.0.0-20200128210451-9ae521aecfd0
	github.com/keys-pub/keysd/http/api v0.0.0-20200214225533-d7890fd4a9a8
	github.com/keys-pub/keysd/http/server v0.0.0-20200214225615-645438f2a2c5
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.4.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keysd/http/api => ../api

// replace github.com/keys-pub/keysd/http/server => ../server
