module github.com/keys-pub/keys-ext/git

go 1.14

require (
	github.com/go-git/go-git/v5 v5.1.0
	github.com/keys-pub/keys v0.0.0-20200602221939-5aac9a3884c2
	github.com/pkg/errors v0.9.1
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.5.1
)

replace github.com/keys-pub/keys => ../../keys
