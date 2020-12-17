module github.com/keys-pub/keys-ext/vault

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/golang/snappy v0.0.2 // indirect
	github.com/keys-pub/keys v0.1.18-0.20201217003538-138e9a15ccb4
	github.com/keys-pub/keys-ext/http/api v0.0.0-20201217184537-bbf6e26a08c5
	github.com/keys-pub/keys-ext/http/client v0.0.0-20201217212639-fc124fb22380
	github.com/keys-pub/keys-ext/http/server v0.0.0-20201217212545-a3de48980037
	github.com/pkg/errors v0.9.1
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.5.1
	github.com/syndtr/goleveldb v1.0.0
	github.com/vmihailenco/msgpack/v4 v4.3.12
	github.com/vmihailenco/tagparser v0.1.2 // indirect
	golang.org/x/crypto v0.0.0-20201217014255-9d1352758620
	google.golang.org/appengine v1.6.7 // indirect
)

// replace github.com/keys-pub/keys => ../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../http/api

// replace github.com/keys-pub/keys-ext/http/client => ../http/client

// replace github.com/keys-pub/keys-ext/http/server => ../http/server
