module github.com/keys-pub/keys-ext/http/client

go 1.14

require (
	github.com/keys-pub/keys v0.1.20
	github.com/keys-pub/keys-ext/firestore v0.0.0-20210306221652-cf68a1890228
	github.com/keys-pub/keys-ext/http/api v0.0.0-20210203191236-2bce35af93a0
	github.com/keys-pub/keys-ext/http/server v0.0.0-20210118231903-89d20ffc493c
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	google.golang.org/api v0.40.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../api

// replace github.com/keys-pub/keys-ext/http/server => ../server

// replace github.com/keys-pub/keys-ext/firestore => ../../firestore

// replace github.com/keys-pub/keys-ext/ws/api => ../../ws/api
