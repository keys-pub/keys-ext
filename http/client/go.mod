module github.com/keys-pub/keysd/http/client

go 1.12

require (
	github.com/keys-pub/keys v0.0.0-20200314221622-6f842a6599a4
	github.com/keys-pub/keysd/http/api v0.0.0-20200314221420-0da6c9407c23
	github.com/keys-pub/keysd/http/server v0.0.0-20200314224442-13ee3e39a493
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.4.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keysd/http/api => ../api

// replace github.com/keys-pub/keysd/http/server => ../server
