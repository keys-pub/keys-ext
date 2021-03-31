module github.com/keys-pub/keys-ext/firestore

go 1.12

require (
	cloud.google.com/go v0.80.0 // indirect
	cloud.google.com/go/firestore v1.5.0
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/keys-pub/keys v0.1.21-0.20210331163518-474087d0d185
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/net v0.0.0-20210331060903-cb1fcc7394e5 // indirect
	golang.org/x/oauth2 v0.0.0-20210323180902-22b0adad7558 // indirect
	google.golang.org/api v0.43.0
	google.golang.org/genproto v0.0.0-20210331142528-b7513248f0ba // indirect
	google.golang.org/grpc v1.36.1
)

// replace github.com/keys-pub/keys => ../../keys
