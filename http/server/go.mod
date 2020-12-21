module github.com/keys-pub/keys-ext/http/server

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/gorilla/websocket v1.4.2
	github.com/keys-pub/keys v0.1.18-0.20201221022455-0512e1665806
	github.com/keys-pub/keys-ext/firestore v0.0.0-20201120035752-fc8566e1f7c4
	github.com/keys-pub/keys-ext/http/api v0.0.0-20201218211059-81db8e866f8c
	github.com/keys-pub/keys-ext/ws/api v0.0.0-20201217184537-bbf6e26a08c5
	github.com/labstack/echo/v4 v4.1.17
	github.com/labstack/gommon v0.3.0
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	github.com/vmihailenco/msgpack/v4 v4.3.12
	golang.org/x/crypto v0.0.0-20201217014255-9d1352758620 // indirect
	golang.org/x/net v0.0.0-20201216054612-986b41b23924 // indirect
	golang.org/x/sys v0.0.0-20201214210602-f9fddec55a1e // indirect
	google.golang.org/api v0.35.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../api

// replace github.com/keys-pub/keys-ext/firestore => ../../firestore

// replace github.com/keys-pub/keys-ext/ws/api => ../../ws/api
