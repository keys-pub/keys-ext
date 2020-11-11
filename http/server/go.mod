module github.com/keys-pub/keys-ext/http/server

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/gorilla/websocket v1.4.2
	github.com/keys-pub/keys v0.1.18-0.20201110225229-cf94f4121589
	github.com/keys-pub/keys-ext/firestore v0.0.0-20201110233325-bd40b7e46e7d
	github.com/keys-pub/keys-ext/http/api v0.0.0-20201111235049-c5301b1972b8
	github.com/labstack/echo/v4 v4.1.17
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	google.golang.org/api v0.35.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../api

// replace github.com/keys-pub/keys-ext/firestore => ../../firestore
