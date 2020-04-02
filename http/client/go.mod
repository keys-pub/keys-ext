module github.com/keys-pub/keysd/http/client

go 1.12

require (
	github.com/gorilla/websocket v1.4.2
	github.com/keys-pub/keys v0.0.0-20200402182809-2e21a424687e
	github.com/keys-pub/keysd/http/api v0.0.0-20200402183018-a85eceb453b1
	github.com/keys-pub/keysd/http/server v0.0.0-20200402191721-a864d6b0e313
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.4.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keysd/http/api => ../api

// replace github.com/keys-pub/keysd/http/server => ../server
