module github.com/keys-pub/keysd/db

go 1.12

require (
	github.com/danieljoos/wincred v1.0.3-0.20190627210546-1fd2f0dfbd6a // indirect
	github.com/gabriel/go-keychain v0.0.0-20191220021328-378d9d7f4318 // indirect
	github.com/gabriel/goleveldb-encrypted v0.0.0-20191220210737-4aefa2aa0d62
	github.com/golang/snappy v0.0.1 // indirect
	github.com/keys-pub/keys v0.0.0-20200124060448-34fed9f6ffa9
	github.com/kr/pretty v0.1.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/syndtr/goleveldb v1.0.0
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

// replace github.com/keys-pub/keys => ../../keys
