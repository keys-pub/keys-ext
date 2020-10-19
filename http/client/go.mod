module github.com/keys-pub/keys-ext/http/client

go 1.14

require (
	github.com/keys-pub/keys v0.1.7-0.20201019222734-27495f7e1624
	github.com/keys-pub/keys-ext/firestore v0.0.0-20200803193547-52c161dbd094
	github.com/keys-pub/keys-ext/http/api v0.0.0-20201019222921-495773d46954
	github.com/keys-pub/keys-ext/http/server v0.0.0-20201019223136-6de853aa20e6
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	github.com/vmihailenco/msgpack/v4 v4.3.12
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a // indirect
	golang.org/x/net v0.0.0-20200904194848-62affa334b73 // indirect
	golang.org/x/sys v0.0.0-20200917073148-efd3b9a0ff20 // indirect
	google.golang.org/api v0.25.0
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../api

// replace github.com/keys-pub/keys-ext/http/server => ../server

// replace github.com/keys-pub/keys-ext/firestore => ../../firestore
