module github.com/keys-pub/keysd/http/client

go 1.12

require (
	github.com/keys-pub/keys v0.0.0-20200107221419-0ed0e891d305
	github.com/keys-pub/keysd/http/api v0.0.0-20200107201206-c0f295622c20
	github.com/keys-pub/keysd/http/server v0.0.0-20200107234415-8888386190c0
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.4.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keysd/http/api => ../api

// replace github.com/keys-pub/keysd/http/server => ../server
