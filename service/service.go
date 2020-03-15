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
	open    bool
	openMtx sync.Mutex

	watchLast *keys.WatchEvent
	watchLn   keys.WatchLn
	watchWg   *sync.WaitGroup
	watchMtx  sync.Mutex
}

func newService(cfg *Config, build Build, auth *auth, req keys.Requestor, nowFn func() time.Time) (*service, error) {
	ks := keys.NewKeystore(auth.keyring)
	db := db.NewDB()
	db.SetTimeNow(nowFn)
	scs := keys.NewSigchainStore(db)
	users, err := keys.NewUserStore(db, scs, req, nowFn)
	if err != nil {
		return nil, err
	}

	remote, err := client.NewClient(cfg.Server(), ks)
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
// If already open, will close and re-open.
func (s *service) Open() error {
	s.openMtx.Lock()
	defer s.openMtx.Unlock()
	if s.open {
		logger.Errorf("Service already open, closing and re-opening...")
		s.close()
	}
	logger.Infof("Opening db...")
	path, err := s.cfg.AppPath(fmt.Sprintf("keys.leveldb"), true)
	if err != nil {
		return err
	}

	// TODO: leveldb encryption?

	if err := s.db.OpenAtPath(path, nil); err != nil {
		return err
	}
	s.startCheck()
	s.open = true

	return nil
}

// Close ...
func (s *service) Close() {
	s.openMtx.Lock()
	defer s.openMtx.Unlock()
	if !s.open {
		logger.Infof("Service already closed")
		return
	}
	s.close()
}

func (s *service) close() {
	s.stopCheck()
	s.watchReqClose()
	logger.Infof("Closing db...")
	s.db.Close()
	s.open = false
}

func (s *service) check(ctx context.Context) {
	logger.Infof("Checking for expired users...")
	ids, err := s.users.Expired(ctx, time.Hour*24)
	if err != nil {
		logger.Warningf("Failed to get expired users: %v", err)
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

func (s *service) stopCheck() {
	if s.closeCh != nil {
		close(s.closeCh)
		s.closeCh = nil
	}
}
