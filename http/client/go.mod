module github.com/keys-pub/keys-ext/http/client

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/keys-pub/keys v0.1.18-0.20201110225229-cf94f4121589
	github.com/keys-pub/keys-ext/firestore v0.0.0-20201110235704-afb7223ac00d
	github.com/keys-pub/keys-ext/http/api v0.0.0-20201111002057-e7a85c338e3d
	github.com/keys-pub/keys-ext/http/server v0.0.0-20201111183311-6042c730fa0c
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
