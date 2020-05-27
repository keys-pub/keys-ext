package service

import (
	"fmt"

	"github.com/keys-pub/keys/keyring"
	"github.com/vmihailenco/msgpack/v4"
)

// authInfo stores metadata and parameters used for a keyring auth.
// Currently mainly stores FIDO2 HMAC secret credential info.
type authInfo struct {
	ID     string `msgpack:"id"`
	AAGUID string `msgpack:"aaguid"`
	Salt   []byte `msgpack:"salt"`
}

func (a *auth) loadInfo() ([]*authInfo, error) {
	st := a.keyring.Store()
	service := a.cfg.KeyringService(st.Name())
	ids, err := st.IDs(service, keyring.WithReservedPrefix("#info-"))
	if err != nil {
		return nil, err
	}
	logger.Debugf("Looking up auth info %v", ids)
	infos := make([]*authInfo, 0, len(ids))
	for _, id := range ids {
		b, err := st.Get(service, id)
		if err != nil {
			return nil, err
		}
		var info authInfo
		if err := msgpack.Unmarshal(b, &info); err != nil {
			return nil, err
		}
		infos = append(infos, &info)
	}
	return infos, nil
}

func (a *auth) saveInfo(info *authInfo) error {
	logger.Debugf("Saving auth info %s", info.ID)
	st := a.keyring.Store()
	service := a.cfg.KeyringService(st.Name())
	krid := fmt.Sprintf("#info-%s", info.ID)
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
	krid := fmt.Sprintf("#info-%s", id)
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
