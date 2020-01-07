package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/db"
	"github.com/keys-pub/keysd/http/client"
)

type service struct {
	cfg    *Config
	build  Build
	auth   *auth
	db     *db.DB
	ks     *keys.Keystore
	remote *client.Client
	scs    keys.SigchainStore
	users  *keys.UserStore
	nowFn  func() time.Time

	watchLast *keys.WatchEvent
	watchLn   keys.WatchLn
	watchWg   *sync.WaitGroup
	watchMtx  sync.Mutex
}

func newService(cfg *Config, build Build, auth *auth, req keys.Requestor, nowFn func() time.Time) (*service, error) {
	ks := keys.NewKeystore()
	ks.SetKeyring(auth.keyring)

	db := db.NewDB()
	db.SetTimeNow(nowFn)
	scs := keys.NewSigchainStore(db)
	users, err := keys.NewUserStore(db, scs, []string{keys.Twitter, keys.Github}, req, nowFn)
	if err != nil {
		return nil, err
	}

	remote, err := client.NewClient(cfg.Server())
	if err != nil {
		return nil, err
	}
	remote.SetTimeNow(nowFn)

	return &service{
		auth:    auth,
		build:   build,
		cfg:     cfg,
		ks:      ks,
		scs:     scs,
		db:      db,
		users:   users,
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
func (s *service) Open(sk keys.SecretKey) error {
	if s.db.IsOpen() {
		logger.Errorf("DB already open, closing first...")
		s.Close()
	}

	logger.Infof("Opening db...")
	path, err := s.cfg.AppPath(fmt.Sprintf("keys.leveldb"), true)
	if err != nil {
		return err
	}
	if err := s.db.OpenAtPath(path, sk, nil); err != nil {
		return err
	}
	return nil
}

// Close ...
func (s *service) Close() {
	s.watchReqClose()
	if s.db != nil {
		s.db.Close()
	}
}
