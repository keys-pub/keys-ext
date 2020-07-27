module github.com/keys-pub/keys-ext/http/client

go 1.14

require (
	github.com/keys-pub/keys v0.1.2-0.20200727004443-4f0acc292c3b
	github.com/keys-pub/keys-ext/http/api v0.0.0-20200721205504-d589cebeca43
	github.com/keys-pub/keys-ext/http/server v0.0.0-20200721205647-86df18234737
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	github.com/vmihailenco/msgpack/v4 v4.3.11
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../api

// replace github.com/keys-pub/keys-ext/http/server => ../server
