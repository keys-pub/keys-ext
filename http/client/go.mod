module github.com/keys-pub/keys-ext/http/client

go 1.14

require (
	github.com/gorilla/websocket v1.4.2
	github.com/keys-pub/keys v0.0.0-20200612010917-0cf3f60778ea
	github.com/keys-pub/keys-ext/http/api v0.0.0-20200612011605-1b8b64293fa0
	github.com/keys-pub/keys-ext/http/server v0.0.0-20200612011944-4dc928f42721
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../api

// replace github.com/keys-pub/keys-ext/http/server => ../server
