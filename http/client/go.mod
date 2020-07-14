module github.com/keys-pub/keys-ext/http/client

go 1.14

require (
	github.com/gorilla/websocket v1.4.2
	github.com/iancoleman/orderedmap v0.0.0-20190318233801-ac98e3ecb4b0
	github.com/keys-pub/keys v0.1.2-0.20200714015424-b54f7f572bd1
	github.com/keys-pub/keys-ext/http/api v0.0.0-20200704211016-ce8ce10a1087
	github.com/keys-pub/keys-ext/http/server v0.0.0-20200714015603-7e7d65871956
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	github.com/vmihailenco/msgpack v4.0.4+incompatible
	github.com/vmihailenco/msgpack/v4 v4.3.11
	github.com/wk8/go-ordered-map v0.1.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../api

// replace github.com/keys-pub/keys-ext/http/server => ../server
