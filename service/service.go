package service

import (
	"context"
	"sync"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/db"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys-ext/syncp"
	"github.com/keys-pub/keys/keyring"
	"github.com/keys-pub/keys/request"
	"github.com/keys-pub/keys/user"
)

type service struct {
	cfg    *Config
	build  Build
	auth   *auth
	db     *db.DB
	remote *client.Client
	scs    keys.SigchainStore
	users  *user.Store
	nowFn  func() time.Time
	kr     *keyring.Keyring
	scfg   syncp.Config

	closeCh chan bool
	open    bool
	openMtx sync.Mutex
}

func newService(cfg *Config, build Build, auth *auth, keyringType string, req request.Requestor, nowFn func() time.Time) (*service, error) {
	logger.Debugf("New service: %s", cfg.AppName())

	kr, scfg, err := newKeyring(cfg, keyringType)
	if err != nil {
		return nil, err
	}

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
		auth:   auth,
		build:  build,
		cfg:    cfg,
		kr:     kr,
		scfg:   scfg,
		scs:    scs,
		db:     db,
		users:  users,
		remote: remote,
		nowFn:  nowFn,
	}, nil
}

// Now ...
func (s *service) Now() time.Time {
	return s.nowFn()
}

const dbFilename = "keys.sdb"

// TODO: Remove old db "keys.leveldb"

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

	exists, err := pathExists(path)
	if err != nil {
		return err
	}
	isNew := false
	if !exists {
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
