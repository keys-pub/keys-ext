module github.com/keys-pub/keysd/http/client

go 1.12

require (
	github.com/keys-pub/keys v0.0.0-20200118211353-b5b965520f79
	github.com/keys-pub/keysd/http/api v0.0.0-20200118211709-e26c534fe6c2
	github.com/keys-pub/keysd/http/server v0.0.0-20200108235830-75282020aeea
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.4.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keysd/http/api => ../api

// replace github.com/keys-pub/keysd/http/server => ../server
