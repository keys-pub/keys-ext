package vault

import (
	"context"
	"net/url"
	"time"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// Remote for vault.
type Remote struct {
	URL  *url.URL          `json:"url,omitempty"`
	Key  *keys.EdX25519Key `json:"key"`
	Salt []byte            `json:"salt"`
}

// NewRemote creates a Remote.
func NewRemote(url *url.URL, key *keys.EdX25519Key, salt []byte) *Remote {
	return &Remote{URL: url, Key: key, Salt: salt}
}

// Clone initializes Vault with from remote.
func (v *Vault) Clone(ctx context.Context, remote *Remote) error {
	if remote == nil {
		return errors.Errorf("nil remote")
	}
	if remote.URL != nil {
		return errors.Errorf("url not supported")
	}
	if v.client == nil {
		return errors.Errorf("no vault client set")
	}

	empty, err := v.IsEmpty()
	if err != nil {
		return err
	}
	if !empty {
		return errors.Errorf("vault not empty, can only be initialized if empty")
	}

	if len(remote.Salt[:]) == 0 {
		return errors.Errorf("no remote salt")
	}

	logger.Infof("Requesting remote vault...")
	vault, err := v.client.Vault(ctx, remote.Key)
	if err != nil {
		return err
	}

	if err := v.setRemoteSalt(remote.Salt[:]); err != nil {
		return err
	}

	if err := v.saveRemoteVault(vault); err != nil {
		return err
	}

	if err := v.setLastSync(time.Now()); err != nil {
		return err
	}

	v.remote = remote
	return nil
}

// Remote returns remote server and auth, if unlocked.
// The vault auth key is used to encrypt and verify vault items from the server.
// This encryption happens on top of the encryption by the master key.
// TODO: Point to spec.
func (v *Vault) Remote() *Remote {
	return v.remote
}
