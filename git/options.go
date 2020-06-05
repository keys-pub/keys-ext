package git

import (
	"time"

	"github.com/go-git/go-git/v5/plumbing/transport"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/keys-pub/keys"
)

// Options ...
type Options struct {
	auth  transport.AuthMethod
	krd   string
	nowFn func() time.Time
}

// Option ...
type Option func(*Options) error

func newOptions(opts ...Option) (Options, error) {
	options := Options{
		nowFn: time.Now,
	}
	for _, o := range opts {
		if err := o(&options); err != nil {
			return options, err
		}
	}
	return options, nil
}

// Key sets the ssh key.
func Key(key *keys.EdX25519Key) Option {
	return func(o *Options) error {
		o.auth = &gitssh.PublicKeys{User: "git", Signer: key.SSHSigner()}
		return nil
	}
}

// KeyringDir if a subdirectory should be used for the keyring.
// Used for testing.
func KeyringDir(krd string) Option {
	return func(o *Options) error {
		o.krd = krd
		return nil
	}
}

// NowFn to set clock.
// Used for testing.
func NowFn(nowFn func() time.Time) Option {
	return func(o *Options) error {
		o.nowFn = nowFn
		return nil
	}
}
