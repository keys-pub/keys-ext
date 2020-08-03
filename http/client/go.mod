module github.com/keys-pub/keys-ext/http/client

go 1.14

require (
	github.com/keys-pub/keys v0.1.2-0.20200731213842-a306c75de40a
	github.com/keys-pub/keys-ext/firestore v0.0.0-20200803193547-52c161dbd094
	github.com/keys-pub/keys-ext/http/api v0.0.0-20200730003632-c95092bc23ed
	github.com/keys-pub/keys-ext/http/server v0.0.0-20200803193934-036dcb4c6007
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	github.com/vmihailenco/msgpack/v4 v4.3.11
	google.golang.org/api v0.25.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../api

// replace github.com/keys-pub/keys-ext/http/server => ../server

// replace github.com/keys-pub/keys-ext/firestore => ../../firestore
