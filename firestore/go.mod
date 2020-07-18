module github.com/keys-pub/keys-ext/firestore

go 1.12

require (
	cloud.google.com/go v0.51.0 // indirect
	cloud.google.com/go/firestore v1.1.0
	github.com/keys-pub/keys v0.1.2-0.20200718011252-5bff924a7b82
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/tools v0.0.0-20200410194907-79a7a3126eef // indirect
	google.golang.org/api v0.15.0
	google.golang.org/grpc v1.26.0
)

// replace github.com/keys-pub/keys => ../../keys
