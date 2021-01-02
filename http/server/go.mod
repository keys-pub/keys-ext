module github.com/keys-pub/keys-ext/http/server

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/gorilla/websocket v1.4.2
	github.com/keys-pub/keys v0.1.20-0.20210102022201-ffb45798b8ab
	github.com/keys-pub/keys-ext/firestore v0.0.0-20210102023225-d2e7279d30fc
	github.com/keys-pub/keys-ext/http/api v0.0.0-20210102234606-16f23aaf0966
	github.com/keys-pub/keys-ext/ws/api v0.0.0-20210102023225-d2e7279d30fc
	github.com/labstack/echo/v4 v4.1.17
	github.com/labstack/gommon v0.3.0
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	github.com/vmihailenco/msgpack/v4 v4.3.12
	google.golang.org/api v0.36.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../api

// replace github.com/keys-pub/keys-ext/firestore => ../../firestore

// replace github.com/keys-pub/keys-ext/ws/api => ../../ws/api
