package service

import (
	"sync"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/saltpack"
	"github.com/keys-pub/keysd/db"
	"github.com/keys-pub/keysd/http/client"
	"github.com/pkg/errors"
)

type service struct {
	cfg    *Config
	build  Build
	auth   *auth
	db     *db.DB
	ks     *keys.Keystore
	local  *keys.CryptoStore
	remote *client.Client
	scs    keys.SigchainStore

	watchLast *keys.WatchEvent
	watchLn   keys.WatchLn
	watchWg   *sync.WaitGroup
	watchMtx  sync.Mutex
}

func newService(cfg *Config, build Build, auth *auth) *service {
	ks := keys.NewKeystore()
	ks.SetKeyring(auth.keyring)
	return &service{
		cfg:     cfg,
		build:   build,
		auth:    auth,
		ks:      ks,
		watchLn: func(e *keys.WatchEvent) {},
	}
}

// Open the service.
func (s *service) Open() error {
	if s.db != nil && s.db.IsOpen() {
		return errors.Errorf("db already open")
	}

	logger.Infof("Opening db...")
	s.db = db.NewDB()
	path, err := s.cfg.AppPath("keys.leveldb", true)
	if err != nil {
		return err
	}
	if err := s.db.OpenAtPath(path); err != nil {
		return err
	}

	cp := saltpack.NewSaltpack(s.ks)
	s.local = keys.NewCryptoStore(s.db, cp)
	s.scs = keys.NewSigchainStore(s.local)
	s.ks.SetSigchainStore(s.scs)

	remote, err := client.NewClient("https://keys.pub", cp)
	if err != nil {
		_ = s.Close()
		return err
	}
	s.SetRemote(remote)

	return nil
}

// SetRemote sets the remote.
func (s *service) SetRemote(remote *client.Client) {
	s.remote = remote
}

// Close ...
func (s *service) Close() error {
	s.ks.SetSigchainStore(nil)
	if s.db != nil {
		s.db.Close()
	}

	s.db = nil
	s.local = nil
	s.scs = nil
	s.watchReqClose()
	s.remote = nil
	return nil
}
