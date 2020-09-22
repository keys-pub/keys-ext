package service

import (
	"github.com/keys-pub/keys-ext/vault"
	"github.com/keys-pub/keys/encoding"
	"github.com/pkg/errors"
)

func provisionPaperKey(vlt *vault.Vault, paperKey string) (*vault.Provision, error) {
	key, err := encoding.PhraseToBytes(paperKey, true)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to provision paper key")
	}
	provision := vault.NewProvision(vault.PaperKeyAuth)
	if err := vlt.Provision(key, provision); err != nil {
		return nil, err
	}
	logger.Infof("Provision (paper key): %s", provision.ID)
	return provision, nil
}

func unlockPaperKey(vlt *vault.Vault, paperKey string) (*vault.Provision, error) {
	key, err := encoding.PhraseToBytes(paperKey, true)
	if err != nil {
		return nil, err
	}
	return vlt.Unlock(key)
}
