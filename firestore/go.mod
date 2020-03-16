module github.com/keys-pub/keysd/firestore

go 1.12

require (
	cloud.google.com/go v0.51.0 // indirect
	cloud.google.com/go/firestore v1.1.0
	github.com/danieljoos/wincred v1.0.3-0.20190627210546-1fd2f0dfbd6a // indirect
	github.com/keys-pub/keys v0.0.0-20200316010352-3204a0fbc383
	github.com/pkg/errors v0.9.1
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.4.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/tools v0.0.0-20200107184032-11e9d9cc0042 // indirect
	google.golang.org/api v0.15.0
	google.golang.org/grpc v1.26.0
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

// replace github.com/keys-pub/keys => ../../keys
