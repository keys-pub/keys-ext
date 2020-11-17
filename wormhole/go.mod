module github.com/keys-pub/keys-ext/wormhole

go 1.14

require (
	github.com/keys-pub/keys v0.1.18-0.20201117233052-2bfb5f4d6161
	github.com/keys-pub/keys-ext/http/api v0.0.0-20201110233325-bd40b7e46e7d
	github.com/keys-pub/keys-ext/http/client v0.0.0-20201111001935-49029cf03cae
	github.com/keys-pub/keys-ext/http/server v0.0.0-20201110235704-afb7223ac00d
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
