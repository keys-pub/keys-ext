module github.com/keys-pub/keys-ext/http/server

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/gorilla/websocket v1.4.2
	github.com/keys-pub/keys v0.1.18-0.20201208204604-e9ae8c5755ae
	github.com/keys-pub/keys-ext/firestore v0.0.0-20201120035752-fc8566e1f7c4
	github.com/keys-pub/keys-ext/http/api v0.0.0-20201209191407-ff734efbdcb7
	github.com/keys-pub/keys-ext/ws/api v0.0.0-20201208014450-6b7883173a82
	github.com/labstack/echo/v4 v4.1.17
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	github.com/vmihailenco/msgpack/v4 v4.3.12
	golang.org/x/crypto v0.0.0-20201208171446-5f87f3452ae9 // indirect
	golang.org/x/net v0.0.0-20201209123823-ac852fbbde11 // indirect
	golang.org/x/sys v0.0.0-20201207223542-d4d67f95c62d // indirect
	google.golang.org/api v0.35.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../api

// replace github.com/keys-pub/keys-ext/firestore => ../../firestore

// replace github.com/keys-pub/keys-ext/ws/api => ../../ws/api
