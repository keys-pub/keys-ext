module github.com/keys-pub/keysd/http/client

go 1.12

require (
	github.com/keys-pub/keys v0.0.0-20200126194730-178e13059b7e
	github.com/keys-pub/keysd/http/api v0.0.0-20200126194816-9555aa1b9e60
	github.com/keys-pub/keysd/http/server v0.0.0-20200126194914-e15709360fbd
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.4.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keysd/http/api => ../api

// replace github.com/keys-pub/keysd/http/server => ../server
