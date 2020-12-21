module github.com/keys-pub/keys-ext/wormhole

go 1.14

require (
	github.com/keys-pub/keys v0.1.18-0.20201221024928-926fad6581ab
	github.com/keys-pub/keys-ext/http/api v0.0.0-20201218211059-81db8e866f8c
	github.com/keys-pub/keys-ext/http/client v0.0.0-20201221025613-72a657ea35c1
	github.com/keys-pub/keys-ext/http/server v0.0.0-20201221022604-418ba635ab03
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
