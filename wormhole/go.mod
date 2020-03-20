module github.com/keys-pub/keysd/wormhole

go 1.13

require (
	github.com/keys-pub/keys v0.0.0-20200320021630-30bfb06feb37
	github.com/keys-pub/keysd/http/client v0.0.0-20200320171528-bd2d6df945a0
	github.com/pion/datachannel v1.4.16 // indirect
	github.com/pion/ice v0.7.10 // indirect
	github.com/pion/logging v0.2.2
	github.com/pion/transport v0.9.2 // indirect
	github.com/pion/webrtc/v2 v2.2.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
)

// replace github.com/keys-pub/keys => ../../keys

replace github.com/keys-pub/keysd/http/client => ../http/client
