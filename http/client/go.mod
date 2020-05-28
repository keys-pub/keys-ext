module github.com/keys-pub/keys-ext/http/client

go 1.14

require (
	github.com/gorilla/websocket v1.4.2
	github.com/keys-pub/keys v0.0.0-20200528181135-18e38db61305
	github.com/keys-pub/keys-ext/http/api v0.0.0-20200528184029-7548f2a0a594
	github.com/keys-pub/keys-ext/http/server v0.0.0-20200528185324-90ced7e635aa
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../api

// replace github.com/keys-pub/keys-ext/http/server => ../server