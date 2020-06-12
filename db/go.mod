module github.com/keys-pub/keys-ext/db

go 1.12

require (
	github.com/golang/snappy v0.0.1 // indirect
	github.com/keys-pub/keys v0.0.0-20200612010917-0cf3f60778ea
	github.com/minio/sio v0.2.1-0.20191008223331-a3e7c367e48e
	github.com/pkg/errors v0.9.1
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.5.1
	github.com/syndtr/goleveldb v1.0.0
	github.com/vmihailenco/msgpack/v4 v4.3.11
)

// replace github.com/keys-pub/keys => ../../keys
