module github.com/keys-pub/keys-ext/ws/client

go 1.15

require (
	github.com/gorilla/websocket v1.4.2
	github.com/keys-pub/keys v0.1.18-0.20201029233150-785ac922181d
	github.com/keys-pub/keys-ext/ws/api v0.0.0-20201030201518-8a43008be509
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	golang.org/x/crypto v0.0.0-20201016220609-9e8e0b390897 // indirect
	golang.org/x/sys v0.0.0-20201029080932-201ba4db2418 // indirect
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/ws/api => ../api
