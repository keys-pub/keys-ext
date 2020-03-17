module github.com/keys-pub/keysd

go 1.12

require (
	github.com/golang/protobuf v1.3.5 // indirect
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/keys-pub/keys v0.0.0-20200317181626-44d21a618612 // indirect
	github.com/keys-pub/keysd/db v0.0.0-20200317224411-cb5c88e6751b // indirect
	github.com/keys-pub/keysd/http/api v0.0.0-20200317224411-cb5c88e6751b // indirect
	github.com/keys-pub/keysd/service v0.0.0-20200317224411-cb5c88e6751b
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	github.com/urfave/cli v1.22.3 // indirect
	golang.org/x/crypto v0.0.0-20200317142112-1b76d66859c6 // indirect
	golang.org/x/sys v0.0.0-20200317113312-5766fd39f98d // indirect
	google.golang.org/genproto v0.0.0-20200317114155-1f3552e48f24 // indirect
	google.golang.org/grpc v1.28.0 // indirect
)

replace github.com/keys-pub/keysd/service => ./service

// replace github.com/keys-pub/keysd/http/api => ./http/api

// replace github.com/keys-pub/keysd/http/client => ./http/client

// replace github.com/keys-pub/keysd/http/server => ./http/server

// replace github.com/keys-pub/keysd/db => ./db

// replace github.com/keys-pub/keys => ../keys
