module github.com/keys-pub/keysd/http/client

go 1.12

require (
	github.com/keys-pub/keys v0.0.0-20191218212024-b35d6b05a135
	github.com/keys-pub/keysd/http/api v0.0.0-20191218205614-5016b6582dfb
	github.com/keys-pub/keysd/http/server v0.0.0-20191218225836-3a17c4a9b7cc
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.4.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keysd/http/api => ../api

// replace github.com/keys-pub/keysd/http/server => ../server
