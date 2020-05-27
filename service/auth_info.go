package service

import (
	"time"

	"github.com/keys-pub/keys/keyring"
	"github.com/vmihailenco/msgpack/v4"
)

type authType string

const (
	passwordAuth        authType = "password"
	fido2HMACSecretAuth authType = "fido2-hmac-secret"
)

const authInfoPrefix = "#auinfo-"

// authInfo stores metadata and parameters used for a keyring auth.
// Currently mainly stores FIDO2 HMAC secret credential info.
type authInfo struct {
	ID        string    `msgpack:"id"`
	Type      authType  `msgpack:"type"`
	AAGUID    string    `msgpack:"aaguid"`
	Salt      []byte    `msgpack:"salt"`
	NoPin     bool      `msgpack:"noPin"`
	CreatedAt time.Time `msgpack:"createdAt"`
}

func (a *auth) loadInfos() ([]*authInfo, error) {
	st := a.keyring.Store()
	service := a.cfg.KeyringService(st.Name())
	ids, err := st.IDs(service, keyring.WithReservedPrefix(authInfoPrefix))
	if err != nil {
		return nil, err
	}
	logger.Debugf("Looking up auth infos %v", ids)
	infos := make([]*authInfo, 0, len(ids))
	for _, id := range ids {
		b, err := st.Get(service, id)
		if err != nil {
			return nil, err
		}
		if b == nil {
			logger.Errorf("Missing auth info for %s", id)
			continue
		}
		var info authInfo
		if err := msgpack.Unmarshal(b, &info); err != nil {
			return nil, err
		}
		infos = append(infos, &info)
	}
	return infos, nil
}

func (a *auth) loadInfo(id string) (*authInfo, error) {
	st := a.keyring.Store()
	service := a.cfg.KeyringService(st.Name())
	logger.Debugf("Looking up auth info %v", id)

	b, err := st.Get(service, authInfoPrefix+id)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, nil
	}
	var info authInfo
	if err := msgpack.Unmarshal(b, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (a *auth) saveInfo(info *authInfo) error {
	logger.Debugf("Saving auth info %s", info.ID)
	st := a.keyring.Store()
	service := a.cfg.KeyringService(st.Name())
	krid := authInfoPrefix + info.ID
	b, err := msgpack.Marshal(info)
	if err != nil {
		return err
	}
	if err := st.Set(service, krid, b); err != nil {
		return err
	}
	return nil
}

func (a *auth) deleteInfo(id string) (bool, error) {
	logger.Debugf("Deleting auth info %s", id)
	st := a.keyring.Store()
	service := a.cfg.KeyringService(st.Name())
	krid := authInfoPrefix + id
	return st.Delete(service, krid)
}

func matchAAGUID(infos []*authInfo, aaguid string) *authInfo {
	for _, info := range infos {
		if info.AAGUID == aaguid {
			return info
		}
	}
	return nil
}

func authTypeToRPC(t authType) AuthType {
	switch t {
	case passwordAuth:
		return PasswordAuth
	case fido2HMACSecretAuth:
		return FIDO2HMACSecretAuth
	default:
		return UnknownAuth
	}
}
