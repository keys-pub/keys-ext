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

	unlocked  bool
	unlockMtx sync.Mutex

	checkMtx      sync.Mutex
	checking      bool
	checkCancelFn func()
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
		auth:          auth,
		build:         build,
		cfg:           cfg,
		scs:           scs,
		db:            db,
		users:         users,
		client:        client,
		vault:         vlt,
		clock:         clock,
		checkCancelFn: func() {},
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
	s.lock()
	if err := s.vault.Close(); err != nil {
		logger.Errorf("Error closing vault: %v", err)
	}
}

func (s *service) unlock(ctx context.Context, secret string, typ AuthType, client string) (string, error) {
	s.unlockMtx.Lock()
	defer s.unlockMtx.Unlock()

	// Unlock auth/vault (get token)
	token, err := s.auth.unlock(ctx, s.vault, secret, typ, client)
	if err != nil {
		return "", err
	}

	isNew := false
	if !s.db.IsOpen() {
		logger.Infof("Opening sdb...")
		path, err := s.cfg.AppPath(sdbFilename, true)
		if err != nil {
			return "", err
		}

		// Open sdb
		exists, err := pathExists(path)
		if err != nil {
			return "", err
		}
		if !exists {
			isNew = true
		}

		// Derive sdb key
		// TODO: Check if key is wrong
		mk := s.vault.MasterKey()
		dbk := keys.Bytes32(keys.HKDFSHA256(mk[:], 32, nil, []byte("keys.pub/ldb")))
		if err := s.db.OpenAtPath(ctx, path, dbk); err != nil {
			return "", err
		}
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
		s.startCheck()
	}()

	return token, nil
}

func (s *service) lock() {
	s.unlockMtx.Lock()
	defer s.unlockMtx.Unlock()

	if !s.unlocked {
		logger.Infof("Service already locked")
		return
	}

	s.checkCancelFn()
	// We should give it little bit of time to finish checking after the cancel
	// otherwise it might error trying to write to a closed database.
	// This wait isn't strictly required but we do it to be nice.
	// TODO: Use a WaitGroup with a timeout or channel
	for i := 0; i < 100; i++ {
		if !s.checking {
			break
		}
		time.Sleep(time.Millisecond * 10)
	}

	s.auth.lock(s.vault)

	logger.Infof("Closing sdb...")
	s.db.Close()
	s.unlocked = false
}

func (s *service) tryCheck(ctx context.Context) {
	s.checkMtx.Lock()
	defer s.checkMtx.Unlock()
	s.checking = true
	defer func() { s.checking = false }()

	if _, err := s.vault.CheckSync(ctx, time.Duration(0)); err != nil {
		logger.Errorf("Failed to check sync: %v", err)
	}
	if ctx.Err() != nil {
		return
	}
	if err := s.checkKeys(ctx); err != nil {
		logger.Errorf("Failed to check keys: %v", err)
	}
}

func (s *service) startCheck() {
	ticker := time.NewTicker(time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	s.checkCancelFn = cancel

	go func() {
		s.tryCheck(ctx)
		for {
			select {
			case <-ticker.C:
				s.tryCheck(ctx)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}
