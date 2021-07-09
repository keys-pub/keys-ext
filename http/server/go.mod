module github.com/keys-pub/keys-ext/http/server

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/keys-pub/keys v0.1.22-0.20210708223433-a34d3ce96fb2
	github.com/keys-pub/keys-ext/firestore v0.0.0-20210523201126-199b37b87949
	github.com/keys-pub/keys-ext/http/api v0.0.0-20210525002154-62c1010a2830
	github.com/labstack/echo/v4 v4.2.1
	github.com/mattn/go-colorable v0.1.8 // indirect
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
