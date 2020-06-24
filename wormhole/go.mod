module github.com/keys-pub/keys-ext/wormhole

go 1.14

require (
	github.com/keys-pub/keys v0.0.0-20200624005746-a27efaeb8455
	github.com/keys-pub/keys-ext/http/api v0.0.0-20200624010329-d03428c649d4
	github.com/keys-pub/keys-ext/http/client v0.0.0-20200624011302-f8ff14b0dd9d
	github.com/keys-pub/keys-ext/http/server v0.0.0-20200624011036-97441deb14df
	github.com/pion/logging v0.2.2
	github.com/pion/sctp v1.7.6
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	google.golang.org/api v0.25.0
	gortc.io/stun v1.22.2
)

// replace github.com/keys-pub/keys => ../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../http/api

// replace github.com/keys-pub/keys-ext/http/client => ../http/client

// replace github.com/keys-pub/keys-ext/http/server => ../http/server
