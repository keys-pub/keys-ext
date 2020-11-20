module github.com/keys-pub/keys-ext/ws/client

go 1.15

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/gorilla/websocket v1.4.2
	github.com/keys-pub/keys v0.1.18-0.20201029233150-785ac922181d
	github.com/keys-pub/keys-ext/ws/api v0.0.0-20201120040216-9bb267d5584f
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/vmihailenco/msgpack/v4 v4.3.12 // indirect
	github.com/vmihailenco/tagparser v0.1.2 // indirect
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b // indirect
	golang.org/x/sys v0.0.0-20201119102817-f84b799fce68 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/ws/api => ../api
