module github.com/keys-pub/keys-ext/ws/server

go 1.15

require (
	github.com/gomodule/redigo v1.8.2
	github.com/gorilla/websocket v1.4.2
	github.com/keys-pub/keys v0.1.18-0.20201029233150-785ac922181d
	github.com/keys-pub/keys-ext/ws/api v0.0.0-20201030201518-8a43008be509
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/ws/api => ../api
