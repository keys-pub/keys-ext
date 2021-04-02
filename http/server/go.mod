module github.com/keys-pub/keys-ext/http/server

go 1.14

require (
	github.com/badoux/checkmail v1.2.1
	github.com/davecgh/go-spew v1.1.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/keys-pub/keys v0.1.21-0.20210402011617-28dedbda9f32
	github.com/keys-pub/keys-ext/firestore v0.0.0-20210331163823-45f2f255ab89
	github.com/keys-pub/keys-ext/http/api v0.0.0-20210401205654-ff14cd298c61
	github.com/keys-pub/keys-ext/ws/api v0.0.0-20210331163823-45f2f255ab89
	github.com/keys-pub/vault v0.0.0-20210331210903-81b7918663ab
	github.com/labstack/echo/v4 v4.2.1
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/vmihailenco/msgpack/v4 v4.3.12
	google.golang.org/api v0.43.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../api

// replace github.com/keys-pub/keys-ext/firestore => ../../firestore

// replace github.com/keys-pub/vault => ../../../vault

// replace github.com/keys-pub/keys-ext/ws/api => ../../ws/api
