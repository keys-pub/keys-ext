module github.com/keys-pub/keysd/wormhole

go 1.13

require (
	github.com/keys-pub/keys v0.0.0-20200320021630-30bfb06feb37
	github.com/keys-pub/keysd/http/client v0.0.0-20200321032707-a15834606910
	github.com/keys-pub/keysd/http/server v0.0.0-20200320024609-6e3fca134965
	github.com/pion/logging v0.2.2
	github.com/pion/sctp v1.7.6
	github.com/pion/transport v0.9.2 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	golang.org/x/crypto v0.0.0-20200128174031-69ecbb4d6d5d // indirect
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a // indirect
	gortc.io/stun v1.22.1
)

// replace github.com/keys-pub/keys => ../../keys

// replace github.com/keys-pub/keysd/http/client => ../http/client
