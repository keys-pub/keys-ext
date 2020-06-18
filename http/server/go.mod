module github.com/keys-pub/keys-ext/http/server

go 1.14

require (
	github.com/gorilla/websocket v1.4.2
	github.com/keys-pub/keys v0.0.0-20200618211112-96955ab2a908
	github.com/keys-pub/keys-ext/firestore v0.0.0-20200612011605-1b8b64293fa0
	github.com/keys-pub/keys-ext/http/api v0.0.0-20200618211624-e8000cad93a4
	github.com/labstack/echo/v4 v4.1.16
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	golang.org/x/net v0.0.0-20200602114024-627f9648deb9 // indirect
	golang.org/x/text v0.3.3 // indirect
	google.golang.org/api v0.25.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../api

// replace github.com/keys-pub/keys-ext/firestore => ../../firestore
