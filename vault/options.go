package vault

import (
	"time"

	"github.com/keys-pub/keys"
)

// Options for Vault.
type Options struct {
	Clock func() time.Time
}

// Option for Vault.
type Option func(*Options)

func newOptions(opts ...Option) Options {
	options := Options{
		Clock: time.Now,
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

// SortDirection direction for sorting.
type SortDirection string

const (
	// Ascending direction.
	Ascending SortDirection = "asc"
	// Descending direction.
	Descending SortDirection = "desc"
)

// SecretsOptions ...
type SecretsOptions struct {
	Query         string
	Sort          string
	SortDirection SortDirection
}

// SecretsOption ...
type SecretsOption func(*SecretsOptions)

func newSecretsOptions(opts ...SecretsOption) SecretsOptions {
	var options SecretsOptions
	for _, o := range opts {
		o(&options)
	}
	return options
}
