package service

import (
	"github.com/keys-pub/keys-ext/vault"
	"github.com/pkg/errors"
)

func newVault(cfg *Config, vaultType string, opt ...vault.Option) (*vault.Vault, error) {
	var st vault.Store
	switch vaultType {
	case "":
		krs, err := newKeyringStore(cfg, "")
		if err != nil {
			return nil, err
		}
		st = krs
	case "mem":
		krs, err := newKeyringStore(cfg, "mem")
		if err != nil {
			return nil, err
		}
		st = krs
	default:
		return nil, errors.Errorf("unknown vault type %s", vaultType)
	}

	vlt := vault.New(st, opt...)
	return vlt, nil
}
