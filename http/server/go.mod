module github.com/keys-pub/keysd/http/server

go 1.12

require (
	github.com/gorilla/websocket v1.4.2
	github.com/keys-pub/keys v0.0.0-20200413002436-33c0c3d4ec1b
	github.com/keys-pub/keysd/firestore v0.0.0-20200413003414-34e8a825f8fd
	github.com/keys-pub/keysd/http/api v0.0.0-20200412190331-0e28c0a8f66f
	github.com/labstack/echo/v4 v4.1.11
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.4.0
	github.com/valyala/fasttemplate v1.1.0 // indirect
	google.golang.org/api v0.20.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keysd/http/api => ../api

// replace github.com/keys-pub/keysd/firestore => ../../firestore
