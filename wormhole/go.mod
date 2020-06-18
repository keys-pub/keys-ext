module github.com/keys-pub/keys-ext/wormhole

go 1.14

require (
	github.com/keys-pub/keys v0.0.0-20200618211112-96955ab2a908
	github.com/keys-pub/keys-ext/http/api v0.0.0-20200618211624-e8000cad93a4
	github.com/keys-pub/keys-ext/http/client v0.0.0-20200618212627-7af53dede812
	github.com/keys-pub/keys-ext/http/server v0.0.0-20200618211917-c3daf3e8ad10
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
