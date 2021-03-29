module github.com/keys-pub/keys-ext/ws/server

go 1.15

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/golang/protobuf v1.5.1 // indirect
	github.com/gomodule/redigo v1.8.4
	github.com/gorilla/websocket v1.4.2
	github.com/joho/godotenv v1.3.0
	github.com/keys-pub/keys v0.1.20
	github.com/keys-pub/keys-ext/ws/api v0.0.0-20210328225610-8f5fb2bf254e
	github.com/pkg/errors v0.9.1
	github.com/vmihailenco/msgpack/v4 v4.3.12 // indirect
	github.com/vmihailenco/tagparser v0.1.2 // indirect
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2 // indirect
	golang.org/x/net v0.0.0-20210326220855-61e056675ecf // indirect
	golang.org/x/sys v0.0.0-20210326220804-49726bf1d181 // indirect
	google.golang.org/appengine v1.6.7 // indirect
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/ws/api => ../api
