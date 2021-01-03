module github.com/keys-pub/keys-ext/http/client

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/keys-pub/keys v0.1.20-0.20210102022201-ffb45798b8ab
	github.com/keys-pub/keys-ext/firestore v0.0.0-20210102023225-d2e7279d30fc
	github.com/keys-pub/keys-ext/http/api v0.0.0-20210103224716-09cb2ea897fb
	github.com/keys-pub/keys-ext/http/server v0.0.0-20210103224830-a8720c3ddd9b
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	github.com/vmihailenco/msgpack/v4 v4.3.12
	google.golang.org/api v0.36.0
	google.golang.org/protobuf v1.25.0 // indirect
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../api

// replace github.com/keys-pub/keys-ext/http/server => ../server

// replace github.com/keys-pub/keys-ext/firestore => ../../firestore

// replace github.com/keys-pub/keys-ext/ws/api => ../../ws/api
