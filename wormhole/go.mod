module github.com/keys-pub/keys-ext/wormhole

go 1.14

require (
	github.com/keys-pub/keys v0.0.0-20200612010917-0cf3f60778ea
	github.com/keys-pub/keys-ext/http/api v0.0.0-20200612011605-1b8b64293fa0
	github.com/keys-pub/keys-ext/http/client v0.0.0-20200612012035-74f6a4bba875
	github.com/keys-pub/keys-ext/http/server v0.0.0-20200612011944-4dc928f42721
	github.com/pion/logging v0.2.2
	github.com/pion/sctp v1.7.6
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	gortc.io/stun v1.22.2
)

// replace github.com/keys-pub/keys => ../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../http/api

// replace github.com/keys-pub/keys-ext/http/client => ../http/client

// replace github.com/keys-pub/keys-ext/http/server => ../http/server
