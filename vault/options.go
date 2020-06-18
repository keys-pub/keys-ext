package vault

import (
	"time"

	"github.com/keys-pub/keys"
)

// Options for Vault.
type Options struct {
	Clock    func() time.Time
	protocol protocol
}

// Option for Vault.
type Option func(*Options)

func newOptions(opts ...Option) Options {
	options := Options{
		Clock:    time.Now,
		protocol: v2{},
	}
	for _, o := range opts {
		o(&options)
	}
	return options
}

// WithClock ...
func WithClock(clock func() time.Time) Option {
	return func(o *Options) {
		o.Clock = clock
	}
}

// KeysOptions ...
type KeysOptions struct {
	Types []keys.KeyType
}

// KeysOption ...
type KeysOption func(*KeysOptions)

func newKeysOptions(opts ...KeysOption) KeysOptions {
	options := KeysOptions{}
	for _, o := range opts {
		o(&options)
	}
	return options
}

// WithKeyTypes ...
func WithKeyTypes(types ...keys.KeyType) KeysOption {
	return func(o *KeysOptions) {
		o.Types = types
	}
}
