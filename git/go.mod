module github.com/keys-pub/keys-ext/git

go 1.14

require (
	github.com/go-git/go-git/v5 v5.1.0
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/keys-pub/git2go v0.0.0-20200529003006-6fe50fc72b35
	github.com/keys-pub/keys v0.0.0-20200602221939-5aac9a3884c2
	github.com/pkg/errors v0.9.1
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.5.1
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37
	golang.org/x/net v0.0.0-20200520182314-0ba52f642ac2 // indirect
)

// replace github.com/keys-pub/keys => ../../keys

// replace github.com/keys-pub/git2go => ../../git2go
