module github.com/keys-pub/keys-ext/http/server

go 1.14

require (
	github.com/badoux/checkmail v1.2.1
	github.com/golang/protobuf v1.5.1 // indirect
	github.com/keys-pub/keys v0.1.21-0.20210326211358-fb3db764000f
	github.com/keys-pub/keys-ext/firestore v0.0.0-20210326150845-39fd96e22101
	github.com/keys-pub/keys-ext/http/api v0.0.0-20210328224815-66d117e28647
	github.com/keys-pub/keys-ext/ws/api v0.0.0-20210328225408-abb780969ed5
	github.com/labstack/echo/v4 v4.2.1
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/vmihailenco/msgpack/v4 v4.3.12
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2 // indirect
	golang.org/x/net v0.0.0-20210326220855-61e056675ecf // indirect
	golang.org/x/sys v0.0.0-20210326220804-49726bf1d181 // indirect
	golang.org/x/term v0.0.0-20201210144234-2321bbc49cbf // indirect
	google.golang.org/api v0.40.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../api

// replace github.com/keys-pub/keys-ext/firestore => ../../firestore

// replace github.com/keys-pub/keys-ext/ws/api => ../../ws/api
