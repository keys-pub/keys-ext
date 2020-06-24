module github.com/keys-pub/keys-ext/vault

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/golang/snappy v0.0.1 // indirect
	github.com/keys-pub/keys v0.0.0-20200624005746-a27efaeb8455
	github.com/keys-pub/keys-ext/http/api v0.0.0-20200624010329-d03428c649d4
	github.com/keys-pub/keys-ext/http/client v0.0.0-20200624011302-f8ff14b0dd9d
	github.com/keys-pub/keys-ext/http/server v0.0.0-20200624011036-97441deb14df
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	github.com/syndtr/goleveldb v1.0.0
	github.com/vmihailenco/msgpack/v4 v4.3.12
	golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/protobuf v1.24.0 // indirect
)

// replace github.com/keys-pub/keys => ../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../http/api

// replace github.com/keys-pub/keys-ext/http/client => ../http/client

// replace github.com/keys-pub/keys-ext/http/server => ../http/server

// replace github.com/keys-pub/keys-ext/sdb => ../sdb
