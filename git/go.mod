module github.com/keys-pub/keysd/git

go 1.14

require (
	github.com/keys-pub/keys v0.0.0-20200519005436-e3845e4fcd35
	github.com/libgit2/git2go/v30 v30.0.3
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
)

replace github.com/keys-pub/keys => ../../keys
