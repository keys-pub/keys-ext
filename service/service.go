package service

import (
	"context"
	"sync"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys-ext/sdb"
	"github.com/keys-pub/keys-ext/vault"
	"github.com/keys-pub/keys/request"
	"github.com/keys-pub/keys/user"
)

type service struct {
	cfg    *Config
	build  Build
	auth   *auth
	db     *sdb.DB
	remote *client.Client
	scs    keys.SigchainStore
	users  *user.Store
	clock  func() time.Time
	vault  *vault.Vault

	stopCh    chan bool
	unlocked  bool
	unlockMtx sync.Mutex
}

// const vdbFilename = "vault.vdb"
const sdbFilename = "keys.sdb"

// TODO: Remove old db "keys.leveldb"

func newService(cfg *Config, build Build, auth *auth, req request.Requestor, vaultType string, clock func() time.Time) (*service, error) {
	logger.Debugf("New service: %s", cfg.AppName())

	db := sdb.New()
	db.SetClock(clock)
	scs := keys.NewSigchainStore(db)
	users, err := user.NewStore(db, scs, req, clock)
	if err != nil {
		return nil, err
	}

	remote, err := client.New(cfg.Server())
	if err != nil {
		return nil, err
	}
	remote.SetClock(clock)

	vlt, err := newVault(cfg, vaultType, vault.V1(), vault.WithClock(clock))
	if err != nil {
		return nil, err
	}
	vlt.SetRemote(remote)

	return &service{
		auth:   auth,
		build:  build,
		cfg:    cfg,
		scs:    scs,
		db:     db,
		users:  users,
		remote: remote,
		vault:  vlt,
		clock:  clock,
	}, nil
}

func (s *service) Open() error {
	// logger.Infof("Opening vault db...")
	// path, err := s.cfg.AppPath(vdbFilename, true)
	// if err != nil {
	// 	return err
	// }
	// if err := s.vault.OpenAtPath(path); err != nil {
	// 	return err
	// }
	return nil
}

func (s *service) Close() {
	s.Lock()
	// vdb.Close()
}

// Unlock the service.
// If already unlocked, will lock and unlock.
func (s *service) Unlock(ctx context.Context, key *[32]byte) error {
	s.unlockMtx.Lock()
	defer s.unlockMtx.Unlock()
	if s.unlocked {
		logger.Errorf("Service already unlocked, re-unlocking...")
		s.lock()
	}
	logger.Infof("Opening sdb...")
	path, err := s.cfg.AppPath(sdbFilename, true)
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

	if err := s.db.OpenAtPath(ctx, path, key); err != nil {
		return err
	}

	s.unlocked = true

	// If database is new, we are either in a new state or from a uninstalled
	// (or migrated) state. In the uninstalled state, we should try to update
	// local db for any keys we have in our vault.
	if isNew {
		if err := s.updateAllKeys(ctx); err != nil {
			logger.Errorf("Failed to update keys on new database: %v", err)
		}
	}

	s.startUpdateCheck()

	return nil
}

// Lock ...
func (s *service) Lock() {
	s.unlockMtx.Lock()
	defer s.unlockMtx.Unlock()
	if !s.unlocked {
		logger.Infof("Service already locked")
		return
	}
	s.lock()
}

func (s *service) lock() {
	s.stopUpdateCheck()
	logger.Infof("Closing db...")
	s.db.Close()
	s.unlocked = false
}

func (s *service) tryCheckForKeyUpdates(ctx context.Context) {
	if err := s.checkForKeyUpdates(ctx); err != nil {
		logger.Errorf("Failed to check keys: %v", err)
	}
}

func (s *service) startUpdateCheck() {
	ticker := time.NewTicker(time.Hour)
	s.stopCh = make(chan bool)
	go func() {
		s.tryCheckForKeyUpdates(context.TODO())
		for {
			select {
			case <-ticker.C:
				s.tryCheckForKeyUpdates(context.TODO())
			case <-s.stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *service) stopUpdateCheck() {
	if s.stopCh != nil {
		close(s.stopCh)
		s.stopCh = nil
	}
}
