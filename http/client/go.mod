module github.com/keys-pub/keysd/http/client

go 1.12

require (
	github.com/gorilla/websocket v1.4.2
	github.com/keys-pub/keys v0.0.0-20200527180456-3546952f005f
	github.com/keys-pub/keysd/http/api v0.0.0-20200527181927-f0409e2de588
	github.com/keys-pub/keysd/http/server v0.0.0-20200527182913-749cd5601e56
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keysd/http/api => ../api

// replace github.com/keys-pub/keysd/http/server => ../server
