package service

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/secret"
	"github.com/keys-pub/keys/user"
	"github.com/keys-pub/keys/util"
	"github.com/keys-pub/keysd/db"
	"github.com/keys-pub/keysd/http/client"
)

// TODO: Detect stale sigchains
// TODO: If db cleared, reload sigchains on startup

type service struct {
	cfg    *Config
	build  Build
	auth   *auth
	db     *db.DB
	ks     *keys.Store
	ss     *secret.Store
	remote *client.Client
	scs    keys.SigchainStore
	users  *user.Store
	nowFn  func() time.Time

	closeCh chan bool
	open    bool
	openMtx sync.Mutex

	fido2 bool

	watchLast *ds.WatchEvent
	watchLn   ds.WatchLn
	watchWg   *sync.WaitGroup
	watchMtx  sync.Mutex
}

func newService(cfg *Config, build Build, auth *auth, req util.Requestor, nowFn func() time.Time) (*service, error) {
	logger.Debugf("New service: %s", cfg.AppName())
	ks := keys.NewStore(auth.keyring)
	ss := secret.NewStore(auth.keyring)
	ss.SetTimeNow(nowFn)
	db := db.New()
	db.SetTimeNow(nowFn)
	scs := keys.NewSigchainStore(db)
	users, err := user.NewStore(db, scs, req, nowFn)
	if err != nil {
		return nil, err
	}

	remote, err := client.New(cfg.Server())
	if err != nil {
		return nil, err
	}
	remote.SetTimeNow(nowFn)

	return &service{
		auth:    auth,
		build:   build,
		cfg:     cfg,
		ks:      ks,
		ss:      ss,
		scs:     scs,
		db:      db,
		users:   users,
		remote:  remote,
		nowFn:   nowFn,
		watchLn: func(e *ds.WatchEvent) {},
	}, nil
}

// Now ...
func (s *service) Now() time.Time {
	return s.nowFn()
}

const dbFilename = "keys.leveldb"

// Open the service.
// If already open, will close and re-open.
func (s *service) Open(ctx context.Context, key *[32]byte) error {
	s.openMtx.Lock()
	defer s.openMtx.Unlock()
	if s.open {
		logger.Errorf("Service already open, closing and re-opening...")
		s.close()
	}
	logger.Infof("Opening db...")
	path, err := s.cfg.AppPath(dbFilename, true)
	if err != nil {
		return err
	}

	isNew := false
	if _, err := os.Stat(path); os.IsNotExist(err) {
		isNew = true
	}

	// TODO: Check if key is wrong.
	if err := s.db.OpenAtPath(ctx, path, key); err != nil {
		return err
	}
	s.open = true

	// If database is new, we are either in a new state or from a uninstalled
	// (or migrated) state. In the uninstalled state, we should try to update
	// local db for any keys we have in our keyring.
	if isNew {
		if err := s.updateAllKeys(ctx); err != nil {
			logger.Errorf("Failed to update keys on new database: %v", err)
		}
	}

	s.startUpdateCheck()

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
	s.stopUpdateCheck()
	s.watchReqClose()
	logger.Infof("Closing db...")
	s.db.Close()
	s.open = false
}

func (s *service) tryCheckForKeyUpdates(ctx context.Context) {
	if err := s.checkForKeyUpdates(ctx); err != nil {
		logger.Errorf("Failed to check keys: %v", err)
	}
}

func (s *service) startUpdateCheck() {
	ticker := time.NewTicker(time.Hour)
	s.closeCh = make(chan bool)
	go func() {
		s.tryCheckForKeyUpdates(context.TODO())
		for {
			select {
			case <-ticker.C:
				s.tryCheckForKeyUpdates(context.TODO())
			case <-s.closeCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *service) stopUpdateCheck() {
	if s.closeCh != nil {
		close(s.closeCh)
		s.closeCh = nil
	}
}
