module github.com/keys-pub/keysd/wormhole

go 1.13

require (
	github.com/keys-pub/keys v0.0.0-20200316013823-95ce7c6cb5fa
	github.com/keys-pub/keysd/http/client v0.0.0-20200316164109-d91c033d5a0d
	github.com/pion/webrtc/v2 v2.2.3
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	gortc.io/stun v1.22.1
)

replace github.com/keys-pub/keys => ../../keys
