module github.com/keys-pub/keys-ext/ws/server

go 1.15

require (
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/gomodule/redigo v1.8.2
	github.com/gorilla/websocket v1.4.2
	github.com/joho/godotenv v1.3.0
	github.com/keys-pub/keys v0.1.18-0.20201029233150-785ac922181d
	github.com/keys-pub/keys-ext/ws/api v0.0.0-20201123214159-d624ad274a49
	github.com/pkg/errors v0.9.1
	github.com/vmihailenco/msgpack/v4 v4.3.12
	github.com/vmihailenco/tagparser v0.1.2 // indirect
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b // indirect
	golang.org/x/sys v0.0.0-20201119102817-f84b799fce68 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/ws/api => ../api
