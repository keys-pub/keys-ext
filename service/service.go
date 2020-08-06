package service

import (
	"context"
	"sync"
	"time"

	"github.com/keys-pub/keys"
	httpclient "github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys-ext/sdb"
	"github.com/keys-pub/keys-ext/vault"
	"github.com/keys-pub/keys/request"
	"github.com/keys-pub/keys/tsutil"
	"github.com/keys-pub/keys/user"
)

type service struct {
	cfg    *Config
	build  Build
	auth   *auth
	db     *sdb.DB
	client *httpclient.Client
	scs    *keys.Sigchains
	users  *user.Users
	clock  tsutil.Clock
	vault  *vault.Vault

	stopCh    chan bool
	unlocked  bool
	unlockMtx sync.Mutex
}

const sdbFilename = "keys.sdb"
const vdbFilename = "vault.vdb"

// TODO: Remove old db "keys.leveldb"

func newService(cfg *Config, build Build, auth *auth, req request.Requestor, clock tsutil.Clock) (*service, error) {
	client, err := httpclient.New(cfg.Server())
	if err != nil {
		return nil, err
	}
	client.SetClock(clock)

	path, err := cfg.AppPath(vdbFilename, true)
	if err != nil {
		return nil, err
	}
	vlt := vault.New(vault.NewDB(path), vault.WithClock(clock))
	vlt.SetClient(client)

	db := sdb.New()
	db.SetClock(clock)
	scs := keys.NewSigchains(db)
	users := user.NewUsers(db, scs, req, clock)

	return &service{
		auth:   auth,
		build:  build,
		cfg:    cfg,
		scs:    scs,
		db:     db,
		users:  users,
		client: client,
		vault:  vlt,
		clock:  clock,
	}, nil
}

func (s *service) Open() error {
	logger.Infof("Opening vault...")
	if err := s.vault.Open(); err != nil {
		return err
	}
	if err := checkKeyringConvert(s.cfg, s.vault); err != nil {
		return err
	}
	return nil
}

func (s *service) Close() {
	logger.Infof("Closing...")
	s.Lock()
	if err := s.vault.Close(); err != nil {
		logger.Errorf("Error closing vault db: %v", err)
	}
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

	go func() {
		if _, err := s.vault.CheckSync(context.Background(), time.Duration(0)); err != nil {
			logger.Errorf("Failed to check sync: %v", err)
		}
		s.startCheckKeys()
	}()

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
	s.stopCheckKeys()
	logger.Infof("Closing sdb...")
	s.db.Close()
	s.unlocked = false
}

func (s *service) tryCheckKeys(ctx context.Context) {
	if err := s.checkKeys(ctx); err != nil {
		logger.Errorf("Failed to check keys: %v", err)
	}
}

func (s *service) startCheckKeys() {
	ticker := time.NewTicker(time.Hour)
	s.stopCh = make(chan bool)
	go func() {
		s.tryCheckKeys(context.TODO())
		for {
			select {
			case <-ticker.C:
				s.tryCheckKeys(context.TODO())
			case <-s.stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *service) stopCheckKeys() {
	if s.stopCh != nil {
		close(s.stopCh)
		s.stopCh = nil
	}
}
