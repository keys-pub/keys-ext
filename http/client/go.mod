module github.com/keys-pub/keysd/http/client

go 1.12

require (
	github.com/keys-pub/keys v0.0.0-20200124060448-34fed9f6ffa9
	github.com/keys-pub/keysd/http/api v0.0.0-20200118233442-784585ea9454
	github.com/keys-pub/keysd/http/server v0.0.0-20200118233608-49fe15b1a394
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.4.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keysd/http/api => ../api

// replace github.com/keys-pub/keysd/http/server => ../server
