module github.com/keys-pub/keys-ext/wormhole

go 1.14

require (
	github.com/keys-pub/keys v0.1.2-0.20200718011252-5bff924a7b82
	github.com/keys-pub/keys-ext/http/api v0.0.0-20200718011453-c9ffd4a59862
	github.com/keys-pub/keys-ext/http/client v0.0.0-20200704211703-900392aae3e8
	github.com/keys-pub/keys-ext/http/server v0.0.0-20200704211557-17fe0a678475
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
