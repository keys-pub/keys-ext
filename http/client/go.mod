module github.com/keys-pub/keysd/http/client

go 1.12

require (
	github.com/gorilla/websocket v1.4.2
	github.com/keys-pub/keys v0.0.0-20200414165426-6b7f7009114b
	github.com/keys-pub/keysd/http/api v0.0.0-20200414165929-c63be6975df3
	github.com/keys-pub/keysd/http/server v0.0.0-20200414170216-f05db9c1b454
	github.com/labstack/echo v3.3.10+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.4.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keysd/http/api => ../api

// replace github.com/keys-pub/keysd/http/server => ../server
