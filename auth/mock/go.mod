module github.com/keys-pub/keys-ext/auth/mock

go 1.14

require (
	github.com/google/uuid v1.1.1
	github.com/keys-pub/keys v0.0.0-20200618211112-96955ab2a908
	github.com/keys-pub/keys-ext/auth/fido2 v0.0.0-20200909204406-3d7d2bdd32f8
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	github.com/vmihailenco/msgpack/v4 v4.3.12 // indirect
	golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9 // indirect
	google.golang.org/appengine v1.6.6 // indirect
)

// replace github.com/keys-pub/keys-ext/auth/fido2 => ../fido2
