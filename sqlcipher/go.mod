module github.com/keys-pub/keys-ext/sqlcipher

go 1.15

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/keys-pub/keys v0.1.19
	github.com/keys-pub/keys-ext/sdb v0.0.0-20210113195955-78cba9e669a9
	github.com/mutecomm/go-sqlcipher/v4 v4.4.2
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/goleveldb v1.0.0
	github.com/vmihailenco/msgpack/v4 v4.3.12
)

replace github.com/keys-pub/keys => ../../keys
