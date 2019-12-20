package service

import (
	"sync"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/saltpack"
	"github.com/keys-pub/keysd/db"
	"github.com/keys-pub/keysd/http/client"
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
	uc     *keys.UserContext
	nowFn  func() time.Time

	watchLast *keys.WatchEvent
	watchLn   keys.WatchLn
	watchWg   *sync.WaitGroup
	watchMtx  sync.Mutex
}

func newService(cfg *Config, build Build, auth *auth, uc *keys.UserContext, nowFn func() time.Time) (*service, error) {
	ks := keys.NewKeystore()
	ks.SetKeyring(auth.keyring)

	remote, err := client.NewClient(cfg.Server(), saltpack.NewSaltpack(ks))
	if err != nil {
		return nil, err
	}
	remote.SetTimeNow(nowFn)

	return &service{
		auth:    auth,
		build:   build,
		cfg:     cfg,
		ks:      ks,
		uc:      uc,
		remote:  remote,
		nowFn:   nowFn,
		watchLn: func(e *keys.WatchEvent) {},
	}, nil
}

// Now ...
func (s *service) Now() time.Time {
	return s.nowFn()
}

// Open the service.
func (s *service) Open(key keys.Key) error {
	if s.db != nil && s.db.IsOpen() {
		logger.Errorf("DB already open, closing first...")
		s.Close()
	}

	logger.Infof("Opening db...")
	s.db = db.NewDB()
	path, err := s.cfg.AppPath("keys.leveldb", true)
	if err != nil {
		return err
	}
	if err := s.db.OpenAtPath(path, key.SecretKey(), nil); err != nil {
		return err
	}
	s.db.SetTimeNow(s.nowFn)

	s.local = keys.NewCryptoStore(s.db, saltpack.NewSaltpack(s.ks))
	s.scs = keys.NewSigchainStore(s.local)
	s.ks.SetSigchainStore(s.scs)

	// Generate a sigchain if we don't have one.
	sc, err := s.scs.Sigchain(key.ID())
	if err != nil {
		return err
	}
	if sc == nil {
		if err := s.scs.SaveSigchain(keys.GenerateSigchain(key, s.Now())); err != nil {
			return err
		}
	}

	return nil
}

// Close ...
func (s *service) Close() {
	s.watchReqClose()
	s.ks.SetSigchainStore(nil)
	if s.db != nil {
		s.db.Close()
	}

	s.db = nil
	s.local = nil
	s.scs = nil
}
