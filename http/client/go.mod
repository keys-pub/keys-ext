module github.com/keys-pub/keysd/http/client

go 1.12

require (
	github.com/keys-pub/keys v0.0.0-20200324163809-892a74504eee
	github.com/keys-pub/keysd/http/api v0.0.0-20200324162244-f1b2f8a71cc6
	github.com/keys-pub/keysd/http/server v0.0.0-20200324170537-0d401cf1cd69
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.4.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keysd/http/api => ../api

// replace github.com/keys-pub/keysd/http/server => ../server
