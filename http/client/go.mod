module github.com/keys-pub/keysd/http/client

go 1.12

require (
	github.com/gorilla/websocket v1.4.2
	github.com/keys-pub/keys v0.0.0-20200413004324-63e8a0774692
	github.com/keys-pub/keysd/http/api v0.0.0-20200412190331-0e28c0a8f66f
	github.com/keys-pub/keysd/http/server v0.0.0-20200413003542-7dbbe8346758
	github.com/labstack/echo v3.3.10+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.4.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keysd/http/api => ../api

// replace github.com/keys-pub/keysd/http/server => ../server
