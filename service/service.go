package service

import (
	"context"
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

	closeCh chan bool

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

	// TODO: leveldb encryption

	if err := s.db.OpenAtPath(path, nil); err != nil {
		return err
	}

	s.startCheck()

	return nil
}

// Close ...
func (s *service) Close() {
	s.watchReqClose()
	if s.db.IsOpen() {
		logger.Infof("Closing db...")
		s.db.Close()
	}
	if s.closeCh != nil {
		close(s.closeCh)
		s.closeCh = nil
	}
}

func (s *service) check(ctx context.Context) {
	logger.Infof("Checking for expired users...")
	ids, err := s.users.Expired(ctx, time.Hour*24)
	if err != nil {
		logger.Errorf("Failed to get expired users: %v", err)
		return
	}
	for _, id := range ids {
		_, err := s.users.Update(ctx, id)
		if err != nil {
			logger.Errorf("Failed to update user index for %s: %v", id, err)
		}
	}
}

func (s *service) startCheck() {
	ticker := time.NewTicker(time.Hour)
	s.closeCh = make(chan bool)
	go func() {
		s.check(context.TODO())
		for {
			select {
			case <-ticker.C:
				s.check(context.TODO())
			case <-s.closeCh:
				ticker.Stop()
				return
			}
		}
	}()
}
