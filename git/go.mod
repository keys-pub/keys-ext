module github.com/keys-pub/keysd/git

go 1.14

require (
	github.com/keys-pub/git2go v0.0.0-20200528045742-d98ed315189d
	github.com/keys-pub/keys v0.0.0-20200527185604-bcec14efcd7b
	github.com/keys-pub/keysd/auth/fido2 v0.0.0-20200527222136-fe3bbef02231
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
)

// replace github.com/keys-pub/keys => ../../keys

// replace github.com/keys-pub/git2go => ../../git2go
