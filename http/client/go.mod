module github.com/keys-pub/keysd/http/client

go 1.12

require (
	github.com/keys-pub/keys v0.0.0-20200306031145-ead068a10f45
	github.com/keys-pub/keysd/http/api v0.0.0-20200223203725-9c5a5d442011
	github.com/keys-pub/keysd/http/server v0.0.0-20200223203834-3dc7040c6558
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.4.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keysd/http/api => ../api

// replace github.com/keys-pub/keysd/http/server => ../server
