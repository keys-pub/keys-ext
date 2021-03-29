module github.com/keys-pub/keys-ext/http/client

go 1.14

require (
	github.com/keys-pub/keys v0.1.21-0.20210326211358-fb3db764000f
	github.com/keys-pub/keys-ext/firestore v0.0.0-20210326150845-39fd96e22101
	github.com/keys-pub/keys-ext/http/api v0.0.0-20210329180739-d34ec0c002a0
	github.com/keys-pub/keys-ext/http/server v0.0.0-20210329181014-5a15d0a4eba1
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	google.golang.org/api v0.40.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../api

// replace github.com/keys-pub/keys-ext/http/server => ../server

// replace github.com/keys-pub/keys-ext/firestore => ../../firestore

// replace github.com/keys-pub/keys-ext/ws/api => ../../ws/api
