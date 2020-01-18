module github.com/keys-pub/keysd/http/server

go 1.12

require (
	github.com/keys-pub/keys v0.0.0-20200118211353-b5b965520f79
	github.com/keys-pub/keysd/http/api v0.0.0-20200107201206-c0f295622c20
	github.com/labstack/echo/v4 v4.1.11
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.4.0
	github.com/valyala/fasttemplate v1.1.0 // indirect
	golang.org/x/net v0.0.0-20191204025024-5ee1b9f4859a // indirect
	golang.org/x/text v0.3.2 // indirect
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keysd/http/api => ../api
