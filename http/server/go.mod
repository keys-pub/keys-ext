module github.com/keys-pub/keysd/http/server

go 1.12

require (
	github.com/danieljoos/wincred v1.0.3-0.20190627210546-1fd2f0dfbd6a // indirect
	github.com/keys-pub/keys v0.0.0-20200324163809-892a74504eee
	github.com/keys-pub/keysd/http/api v0.0.0-20200324205758-903123ffbef9
	github.com/kr/pretty v0.1.0 // indirect
	github.com/labstack/echo/v4 v4.1.11
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/valyala/fasttemplate v1.1.0 // indirect
	golang.org/x/net v0.0.0-20191209160850-c0dbc17a3553
	golang.org/x/text v0.3.2 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keysd/http/api => ../api
