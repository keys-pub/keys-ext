module github.com/keys-pub/keys-ext/auth/mock

go 1.14

require (
	github.com/google/uuid v1.1.2
	github.com/keys-pub/keys v0.1.17
	github.com/keys-pub/keys-ext/auth/fido2 v0.0.0-20201124171340-f41427119d82
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	github.com/tyler-smith/go-bip39 v1.1.0 // indirect
	github.com/vmihailenco/msgpack/v4 v4.3.12 // indirect
	golang.org/x/crypto v0.0.0-20201117144127-c1f2f97bffc9 // indirect
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b // indirect
	golang.org/x/sys v0.0.0-20201119102817-f84b799fce68 // indirect
	golang.org/x/text v0.3.4 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/genproto v0.0.0-20201119123407-9b1e624d6bc4 // indirect
	google.golang.org/grpc v1.33.2 // indirect
)

// replace github.com/keys-pub/keys-ext/auth/fido2 => ../fido2
