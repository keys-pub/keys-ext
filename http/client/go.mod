module github.com/keys-pub/keysd/http/client

go 1.12

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/gorilla/websocket v1.4.2
	github.com/keys-pub/keys v0.0.0-20200324163809-892a74504eee
	github.com/keys-pub/keysd/http/api v0.0.0-20200324205758-903123ffbef9
	github.com/keys-pub/keysd/http/server v0.0.0-20200325185544-08fde3c939ab
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.4.0
	golang.org/x/net v0.0.0-20191209160850-c0dbc17a3553
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keysd/http/api => ../api

replace github.com/keys-pub/keysd/http/server => ../server
