module github.com/keys-pub/keys-ext/http/client

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/keys-pub/keys v0.1.18-0.20201203191616-7881134d0f3d
	github.com/keys-pub/keys-ext/firestore v0.0.0-20201120035752-fc8566e1f7c4
	github.com/keys-pub/keys-ext/http/api v0.0.0-20201203191937-c249020a7399
	github.com/keys-pub/keys-ext/http/server v0.0.0-20201120215828-874010c80395
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	github.com/vmihailenco/msgpack/v4 v4.3.12
	golang.org/x/crypto v0.0.0-20201203163018-be400aefbc4c // indirect
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb // indirect
	golang.org/x/sys v0.0.0-20201202213521-69691e467435 // indirect
	google.golang.org/api v0.35.0
	google.golang.org/protobuf v1.25.0 // indirect
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../api

// replace github.com/keys-pub/keys-ext/http/server => ../server

// replace github.com/keys-pub/keys-ext/firestore => ../../firestore

// replace github.com/keys-pub/keys-ext/ws/api => ../../ws/api
