module github.com/keys-pub/keys-ext/vault

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/golang/snappy v0.0.2 // indirect
	github.com/keys-pub/keys v0.1.18-0.20201203221123-07f825c0677a
	github.com/keys-pub/keys-ext/http/api v0.0.0-20201203191937-c249020a7399
	github.com/keys-pub/keys-ext/http/client v0.0.0-20201120220010-3d9f67cb9121
	github.com/keys-pub/keys-ext/http/server v0.0.0-20201120215828-874010c80395
	github.com/pkg/errors v0.9.1
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.5.1
	github.com/syndtr/goleveldb v1.0.0
	github.com/vmihailenco/msgpack/v4 v4.3.12
	github.com/vmihailenco/tagparser v0.1.2 // indirect
	golang.org/x/crypto v0.0.0-20201124201722-c8d3bf9c5392
	google.golang.org/appengine v1.6.7 // indirect
)

// replace github.com/keys-pub/keys => ../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../http/api

// replace github.com/keys-pub/keys-ext/http/client => ../http/client

// replace github.com/keys-pub/keys-ext/http/server => ../http/server
