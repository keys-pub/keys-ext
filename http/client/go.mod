module github.com/keys-pub/keysd/http/client

go 1.12

require (
	github.com/keys-pub/keys v0.0.0-20200320021630-30bfb06feb37
	github.com/keys-pub/keysd/http/api v0.0.0-20200321022921-59e50adf15e4
	github.com/keys-pub/keysd/http/server v0.0.0-20200320024609-6e3fca134965
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.4.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keysd/http/api => ../api

// replace github.com/keys-pub/keysd/http/server => ../server
