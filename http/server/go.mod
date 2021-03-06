module github.com/keys-pub/keys-ext/http/server

go 1.14

require (
	github.com/badoux/checkmail v1.2.1
	github.com/keys-pub/keys v0.1.20
	github.com/keys-pub/keys-ext/firestore v0.0.0-20210306221652-cf68a1890228
	github.com/keys-pub/keys-ext/http/api v0.0.0-20210203191236-2bce35af93a0
	github.com/keys-pub/keys-ext/ws/api v0.0.0-20210118231903-89d20ffc493c
	github.com/labstack/echo/v4 v4.1.17
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/vmihailenco/msgpack/v4 v4.3.12
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 // indirect
	golang.org/x/exp v0.0.0-20201008143054-e3b2a7f2fdc7 // indirect
	golang.org/x/mod v0.4.1 // indirect
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110 // indirect
	golang.org/x/sys v0.0.0-20210305230114-8fe3ee5dd75b // indirect
	golang.org/x/term v0.0.0-20201210144234-2321bbc49cbf // indirect
	golang.org/x/text v0.3.5 // indirect
	golang.org/x/tools v0.1.0 // indirect
	google.golang.org/api v0.40.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	honnef.co/go/tools v0.0.1-2020.1.6 // indirect
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../api

// replace github.com/keys-pub/keys-ext/firestore => ../../firestore

// replace github.com/keys-pub/keys-ext/ws/api => ../../ws/api
