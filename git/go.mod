module github.com/keys-pub/keysd/git

go 1.14

require (
	github.com/keys-pub/keys v0.0.0-20200518022440-9cb382be570a
	github.com/libgit2/git2go/v30 v30.0.3
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
)

replace github.com/keys-pub/keys => ../../keys
