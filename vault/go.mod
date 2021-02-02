module github.com/keys-pub/keys-ext/vault

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/golang/snappy v0.0.2 // indirect
	github.com/keys-pub/keys v0.1.20-0.20210102022201-ffb45798b8ab
	github.com/keys-pub/keys-ext/http/api v0.0.0-20210118231903-89d20ffc493c
	github.com/keys-pub/keys-ext/http/client v0.0.0-20210202031128-4320eac1c167
	github.com/keys-pub/keys-ext/http/server v0.0.0-20210202025052-528710151063
	github.com/pkg/errors v0.9.1
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/goleveldb v1.0.0
	github.com/vmihailenco/msgpack/v4 v4.3.12
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
)

// replace github.com/keys-pub/keys => ../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../http/api

// replace github.com/keys-pub/keys-ext/http/client => ../http/client

// replace github.com/keys-pub/keys-ext/http/server => ../http/server
