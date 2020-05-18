module github.com/keys-pub/keysd/git

go 1.14

require (
	github.com/keys-pub/keys v0.0.0-20200517232941-a94d2050e2ae
	github.com/libgit2/git2go/v30 v30.0.3
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	github.com/vmihailenco/msgpack/v4 v4.3.11
	github.com/zalando/go-keyring v0.0.0-20200121091418-667557018717
)

// replace github.com/keys-pub/keys => ../../keys
