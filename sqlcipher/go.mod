module github.com/keys-pub/keys-ext/sqlcipher

go 1.15

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/keys-pub/keys v0.1.20-0.20210117205459-b7546f4ef310
	github.com/keys-pub/keys-ext/sdb v0.0.0-20210113195955-78cba9e669a9
	github.com/mutecomm/go-sqlcipher/v4 v4.4.2
	github.com/nbutton23/zxcvbn-go v0.0.0-20201221231540-e56b841a3c88 // indirect
	github.com/pkg/errors v0.9.1
	github.com/securego/gosec v0.0.0-20200401082031-e946c8c39989 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/goleveldb v1.0.0
	github.com/vmihailenco/msgpack v4.0.4+incompatible
	github.com/vmihailenco/msgpack/v4 v4.3.12
	golang.org/x/mod v0.4.1 // indirect
	golang.org/x/tools v0.0.0-20210115202250-e0d201561e39 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

// replace github.com/keys-pub/keys => ../../keys
