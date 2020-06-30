module github.com/keys-pub/keys-ext/vault

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/golang/snappy v0.0.1 // indirect
	github.com/keys-pub/keys v0.0.0-20200627005308-6cb6f26a30a7
	github.com/keys-pub/keys-ext/http/api v0.0.0-20200625223334-74da599991bf
	github.com/keys-pub/keys-ext/http/client v0.0.0-20200625224357-38b58050ef02
	github.com/keys-pub/keys-ext/http/server v0.0.0-20200627005809-4b6e75bb8abb
	github.com/keys-pub/keys-ext/sdb v0.0.0-20200627005809-4b6e75bb8abb // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	github.com/syndtr/goleveldb v1.0.0
	github.com/vmihailenco/msgpack/v4 v4.3.12
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/protobuf v1.24.0 // indirect
)

replace github.com/keys-pub/keys => ../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../http/api

replace github.com/keys-pub/keys-ext/http/client => ../http/client

replace github.com/keys-pub/keys-ext/http/server => ../http/server

// replace github.com/keys-pub/keys-ext/sdb => ../sdb
