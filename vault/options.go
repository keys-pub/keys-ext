package vault

import (
	"github.com/keys-pub/keys/tsutil"
)

// Options for Vault.
type Options struct {
	Clock tsutil.Clock
}

// Option for Vault.
type Option func(*Options)

func newOptions(opts ...Option) Options {
	options := Options{
		Clock: tsutil.NewClock(),
	}
	for _, o := range opts {
		o(&options)
	}
	return options
}

// WithClock ...
func WithClock(clock tsutil.Clock) Option {
	return func(o *Options) {
		o.Clock = clock
	}
}
