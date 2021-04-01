module github.com/keys-pub/keys-ext/http/client

go 1.14

require (
	github.com/keys-pub/keys v0.1.21-0.20210401205743-1f5d721f42c0
	github.com/keys-pub/keys-ext/http/api v0.0.0-20210401205654-ff14cd298c61
	github.com/keys-pub/keys-ext/http/server v0.0.0-20210401205934-8b752a983cd9
	github.com/keys-pub/vault v0.0.0-20210331212657-57f7305326e7
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/vmihailenco/msgpack v4.0.4+incompatible
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../api

// replace github.com/keys-pub/keys-ext/http/server => ../server

// replace github.com/keys-pub/vault => ../../../vault

// replace github.com/keys-pub/keys-ext/firestore => ../../firestore

// replace github.com/keys-pub/keys-ext/ws/api => ../../ws/api
