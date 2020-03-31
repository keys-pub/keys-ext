module github.com/keys-pub/keysd/http/client

go 1.12

require (
	github.com/gorilla/websocket v1.4.2
	github.com/keys-pub/keys v0.0.0-20200331185030-55a865de2063
	github.com/keys-pub/keysd/http/api v0.0.0-20200326173323-d53753a929e2
	github.com/keys-pub/keysd/http/server v0.0.0-20200326173537-f15ed1afc333
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.4.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keysd/http/api => ../api

// replace github.com/keys-pub/keysd/http/server => ../server
