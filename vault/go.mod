module github.com/keys-pub/keys-ext/vault

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/golang/snappy v0.0.1 // indirect
	github.com/keys-pub/keys v0.1.2-0.20200718011252-5bff924a7b82
	github.com/keys-pub/keys-ext/http/api v0.0.0-20200718011453-c9ffd4a59862 // indirect
	github.com/keys-pub/keys-ext/http/client v0.0.0-20200716194119-b796152d6e47
	github.com/keys-pub/keys-ext/http/server v0.0.0-20200714015603-7e7d65871956
	github.com/mattn/go-sqlite3 v1.14.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	github.com/syndtr/goleveldb v1.0.0
	github.com/vmihailenco/msgpack/v4 v4.3.12
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/protobuf v1.24.0 // indirect
)

// replace github.com/keys-pub/keys => ../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../http/api

// replace github.com/keys-pub/keys-ext/http/client => ../http/client

// replace github.com/keys-pub/keys-ext/http/server => ../http/server

// replace github.com/keys-pub/keys-ext/sdb => ../sdb
