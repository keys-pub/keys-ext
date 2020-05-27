module github.com/keys-pub/keysd/http/server

go 1.12

require (
	9fans.net/go v0.0.2 // indirect
	github.com/gorilla/websocket v1.4.2
	github.com/keys-pub/keys v0.0.0-20200527180456-3546952f005f
	github.com/keys-pub/keysd/firestore v0.0.0-20200527181815-3e81575df31a
	github.com/keys-pub/keysd/http/api v0.0.0-20200527181927-f0409e2de588
	github.com/keys-pub/wincred v1.0.4 // indirect
	github.com/labstack/echo/v4 v4.1.11
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/pkg/errors v0.9.1
	github.com/rogpeppe/godef v1.1.2 // indirect
	github.com/stretchr/testify v1.5.1
	github.com/valyala/fasttemplate v1.1.0 // indirect
	google.golang.org/api v0.20.0
)

replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keysd/http/api => ../api

// replace github.com/keys-pub/keysd/firestore => ../../firestore
