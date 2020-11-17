module github.com/keys-pub/keys-ext/http/client

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/keys-pub/keys v0.1.18-0.20201117233052-2bfb5f4d6161
	github.com/keys-pub/keys-ext/firestore v0.0.0-20201117233622-e7ee764fc003
	github.com/keys-pub/keys-ext/http/api v0.0.0-20201117233622-e7ee764fc003
	github.com/keys-pub/keys-ext/http/server v0.0.0-20201117233622-e7ee764fc003
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	github.com/vmihailenco/msgpack/v4 v4.3.12
	golang.org/x/crypto v0.0.0-20201117144127-c1f2f97bffc9 // indirect
	golang.org/x/sys v0.0.0-20201117222635-ba5294a509c7 // indirect
	google.golang.org/api v0.35.0
	google.golang.org/protobuf v1.25.0 // indirect
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../api

// replace github.com/keys-pub/keys-ext/http/server => ../server

// replace github.com/keys-pub/keys-ext/firestore => ../../firestore
