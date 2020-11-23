module github.com/keys-pub/keys-ext/http/server

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/gorilla/websocket v1.4.2
	github.com/keys-pub/keys v0.1.18-0.20201120035008-acb3bbba9752
	github.com/keys-pub/keys-ext/firestore v0.0.0-20201120035752-fc8566e1f7c4
	github.com/keys-pub/keys-ext/http/api v0.0.0-20201120215311-661239608411
	github.com/keys-pub/keys-ext/ws/api v0.0.0-20201123214159-d624ad274a49
	github.com/labstack/echo/v4 v4.1.17
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	github.com/vmihailenco/msgpack/v4 v4.3.12
	google.golang.org/api v0.35.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../api

// replace github.com/keys-pub/keys-ext/firestore => ../../firestore

// replace github.com/keys-pub/keys-ext/ws/api => ../../ws/api
