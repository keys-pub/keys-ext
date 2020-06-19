package vault

import (
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/ds"
	"github.com/pkg/errors"
)

// ErrInvalidAuth if auth is invalid.
var ErrInvalidAuth = errors.New("invalid vault auth")

// ErrLocked if no vault key is set.
var ErrLocked = errors.New("vault is locked")

// ErrAlreadySetup if already setup.
var ErrAlreadySetup = errors.New("vault is already setup")

// Salt is default salt value, generated on first access and persisted
// until Reset().
// This salt value is not encrypted.
// Doesn't require Unlock().
func (v *Vault) Salt() ([]byte, error) {
	cpath := v.protocol.Path(configEntity, "salt")
	salt, err := v.store.Get(cpath)
	if err != nil {
		return nil, err
	}
	if salt == nil {
		salt = keys.Rand32()[:]
		if err := v.store.Set(cpath, salt); err != nil {
			return nil, err
		}
	}
	return salt, nil
}

// AuthType describes an auth method.
type AuthType string

const (
	// UnknownAuth ...
	UnknownAuth AuthType = ""
	// PasswordAuth ...
	PasswordAuth AuthType = "password"
	// FIDO2HMACSecretAuth ...
	FIDO2HMACSecretAuth AuthType = "fido2-hmac-secret" // #nosec
)

// Status for vault.
type Status string

const (
	// Unknown status.
	Unknown Status = ""
	// Setup if setup needed.
	Setup Status = "setup"
	// Unlocked if unlocked.
	Unlocked Status = "unlocked"
	// Locked if locked.
	Locked Status = "locked"
)

// Status returns vault status.
// If there are no auths or provisions, returns vault.Setup.
// Doesn't require Unlock().
// TODO: We may want to re-think hardware provisioning requiring seperate Unlock step on setup.
func (v *Vault) Status() (Status, error) {
	if v.mk != nil {
		return Unlocked, nil
	}
	authed, err := v.hasAuth()
	if err != nil {
		return Unknown, err
	}
	if !authed {
		return Setup, nil
	}
	return Locked, nil
}

// Setup auth, if no auth exists.
// Returns ErrAlreadySetup if already setup.
// Doesn't require Unlock().
func (v *Vault) Setup(key *[32]byte, provision *Provision) error {
	status, err := v.Status()
	if err != nil {
		return err
	}
	if status != Setup {
		return ErrAlreadySetup
	}
	if provision == nil {
		return errors.Errorf("no provision")
	}
	mk, err := v.authSetup(provision.ID, key)
	if err != nil {
		return err
	}

	if provision != nil {
		if err := v.saveProvision(provision); err != nil {
			return err
		}
	}

	v.mk = mk
	return nil
}

// Unlock with auth.
// Returns provision used to unlock.
func (v *Vault) Unlock(key *[32]byte) (*Provision, error) {
	logger.Infof("Unlocking...")
	id, mk, err := v.authUnlock(key)
	if err != nil {
		return nil, err
	}
	if mk == nil {
		return nil, ErrInvalidAuth
	}
	logger.Infof("Unlocked with %s", id)
	v.mk = mk

	provision, err := v.loadProvision(id)
	if err != nil {
		return nil, err
	}
	if provision == nil {
		provision = &Provision{ID: id}
	}
	return provision, nil
}

// Lock the vault.
func (v *Vault) Lock() error {
	v.mk = nil
	return nil
}

// authSetup creates master key and encrypts it with the auth key.
func (v *Vault) authSetup(id string, key *[32]byte) (*[32]byte, error) {
	if id == "" {
		return nil, errors.Errorf("no auth id")
	}
	// MK is the master key, setup creates it.
	mk := keys.Rand32()
	if err := v.authCreate(id, key, mk); err != nil {
		return nil, err
	}
	return mk, nil
}

// authCreate encrypts master key with auth key.
func (v *Vault) authCreate(id string, key *[32]byte, mk *[32]byte) error {
	if mk == nil {
		return ErrLocked
	}
	if id == "" {
		return errors.Errorf("no auth id")
	}
	item := NewItem(id, mk[:], "", time.Now())
	b, err := encryptItem(item, key)
	if err != nil {
		return err
	}
	path := v.protocol.Path(authEntity, id)
	if err := v.store.Set(path, b); err != nil {
		return err
	}
	return nil
}

// authDelete removes an auth key.
func (v *Vault) authDelete(id string) (bool, error) {
	if id == "" {
		return false, errors.Errorf("no auth id")
	}
	path := v.protocol.Path(authEntity, id)
	exists, err := v.store.Exists(path)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}
	return v.store.Delete(path)
}

func (v *Vault) hasAuth() (bool, error) {
	path := v.protocol.Path(authEntity)
	iter, err := v.store.Documents(ds.Prefix(path), ds.NoData())
	if err != nil {
		return false, err
	}
	defer iter.Release()
	doc, err := iter.Next()
	if err != nil {
		return false, err
	}
	if doc == nil {
		return false, nil
	}
	return true, nil
}

// authUnlock returns (id, master key) or ("", nil) if a matching auth
// is not found.
// Auth is found by trying to decrypt auth until one is successful.
func (v *Vault) authUnlock(key *[32]byte) (string, *[32]byte, error) {
	path := v.protocol.Path(authEntity)
	iter, err := v.store.Documents(ds.Prefix(path))
	if err != nil {
		return "", nil, err
	}
	defer iter.Release()
	for {
		doc, err := iter.Next()
		if err != nil {
			return "", nil, err
		}
		if doc == nil {
			break
		}
		logger.Debugf("Trying %s", doc.Path)
		item, err := decryptItem(doc.Data, key)
		if err != nil {
			continue
		}
		if item == nil {
			continue
		}
		if len(item.Data) != 32 {
			continue
		}
		return item.ID, keys.Bytes32(item.Data), nil
	}
	return "", nil, nil
}

// UnlockWithPassword unlocks with a password.
// If setup is true, we are setting up the auth for the first time.
// This is a convenience method, calling Setup or Unlock with KeyForPassword using the Salt.
func (v *Vault) UnlockWithPassword(password string, setup bool) error {
	if password == "" {
		return errors.Errorf("empty password")
	}
	salt, err := v.Salt()
	if err != nil {
		return err
	}
	key, err := keys.KeyForPassword(password, salt)
	if err != nil {
		return err
	}
	if setup {
		provision := NewProvision(PasswordAuth)
		if err := v.Setup(key, provision); err != nil {
			return err
		}
		return nil
	}

	if _, err := v.Unlock(key); err != nil {
		return err
	}
	return nil
}
