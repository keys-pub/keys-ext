module github.com/keys-pub/keysd/wormhole

go 1.13

require (
	github.com/keybase/go-keychain v0.0.0-20200218013740-86d4642e4ce2 // indirect
	github.com/keybase/saltpack v0.0.0-20200228190633-d75baa96bffb // indirect
	github.com/keys-pub/keys v0.0.0-20200317181626-44d21a618612
	github.com/keys-pub/keysd/http/api v0.0.0-20200317224602-68134b1264db // indirect
	github.com/keys-pub/keysd/http/client v0.0.0-20200317224602-68134b1264db
	github.com/keys-pub/keysd/http/server v0.0.0-20200317222721-717bf70f4f22
	github.com/pion/logging v0.2.2
	github.com/pion/quic v0.1.1
	github.com/pion/webrtc/v2 v2.2.3
	github.com/pkg/errors v0.9.1
	github.com/schollz/logger v1.2.0
	github.com/stretchr/testify v1.5.1
	golang.org/x/crypto v0.0.0-20200317142112-1b76d66859c6 // indirect
	golang.org/x/sys v0.0.0-20200317113312-5766fd39f98d // indirect
	gortc.io/stun v1.22.1
)

// replace github.com/keys-pub/keys => ../../keys

// replace github.com/keys-pub/keysd/http/client => ../http/client
