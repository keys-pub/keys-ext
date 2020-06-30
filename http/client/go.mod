module github.com/keys-pub/keys-ext/http/client

go 1.14

require (
	github.com/gorilla/websocket v1.4.2
	github.com/iancoleman/orderedmap v0.0.0-20190318233801-ac98e3ecb4b0
	github.com/keys-pub/keys v0.0.0-20200630190925-2c92ea404bb9
	github.com/keys-pub/keys-ext/http/api v0.0.0-20200625223334-74da599991bf
	github.com/keys-pub/keys-ext/http/server v0.0.0-20200630163632-e64cbb9bc1c1
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	github.com/vmihailenco/msgpack v4.0.4+incompatible
	github.com/vmihailenco/msgpack/v4 v4.3.11
	github.com/wk8/go-ordered-map v0.1.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../api

// replace github.com/keys-pub/keys-ext/http/server => ../server
