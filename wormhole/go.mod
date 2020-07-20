module github.com/keys-pub/keys-ext/wormhole

go 1.14

require (
	github.com/keys-pub/keys v0.1.2-0.20200720190123-59e415d170b7
	github.com/keys-pub/keys-ext/http/api v0.0.0-20200720192509-db9723318640
	github.com/keys-pub/keys-ext/http/client v0.0.0-20200720192843-3c8c1e04cd87
	github.com/keys-pub/keys-ext/http/server v0.0.0-20200720192738-5f31fdd89b88
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
