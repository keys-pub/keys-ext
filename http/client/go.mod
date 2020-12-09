module github.com/keys-pub/keys-ext/http/client

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/keys-pub/keys v0.1.18-0.20201208204604-e9ae8c5755ae
	github.com/keys-pub/keys-ext/firestore v0.0.0-20201120035752-fc8566e1f7c4
	github.com/keys-pub/keys-ext/http/api v0.0.0-20201209191407-ff734efbdcb7
	github.com/keys-pub/keys-ext/http/server v0.0.0-20201209191701-aea584fd18fb
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	github.com/vmihailenco/msgpack/v4 v4.3.12
	google.golang.org/api v0.35.0
	google.golang.org/protobuf v1.25.0 // indirect
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../api

// replace github.com/keys-pub/keys-ext/http/server => ../server

// replace github.com/keys-pub/keys-ext/firestore => ../../firestore

// replace github.com/keys-pub/keys-ext/ws/api => ../../ws/api
