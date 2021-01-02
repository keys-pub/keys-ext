package service

import (
	"context"
	"sync"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys-ext/sdb"
	"github.com/keys-pub/keys-ext/vault"
	khttp "github.com/keys-pub/keys/http"
	"github.com/keys-pub/keys/tsutil"
	"github.com/keys-pub/keys/users"
)

type service struct {
	UnimplementedKeysServer

	env    *Env
	build  Build
	auth   *auth
	db     *sdb.DB
	client *client.Client
	scs    *keys.Sigchains
	users  *users.Users
	clock  tsutil.Clock
	vault  *vault.Vault

	unlocked  bool
	unlockMtx sync.Mutex

	checkMtx      sync.Mutex
	checking      bool
	checkCancelFn func()
}

const cdbPath = "cache.sdb"
const vdbPath = "vault.vdb"

func newService(env *Env, build Build, auth *auth, hclient khttp.Client, clock tsutil.Clock) (*service, error) {
	client, err := client.New(env.Server())
	if err != nil {
		return nil, err
	}
	client.SetClock(clock)

	path, err := env.AppPath(vdbPath, true)
	if err != nil {
		return nil, err
	}
	vlt := vault.New(vault.NewDB(path), vault.WithClock(clock))
	vlt.SetClient(client)

	db := sdb.New()
	db.SetClock(clock)
	scs := keys.NewSigchains(db)
	usrs := users.New(db, scs, users.Client(hclient), users.Clock(clock))

	return &service{
		auth:          auth,
		build:         build,
		env:           env,
		scs:           scs,
		db:            db,
		users:         usrs,
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

	// TODO: We can remove this soon (for old keyring conversions)...
	if err := checkKeyringConvert(s.env, s.vault); err != nil {
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

// TODO: unlock can be called multiple times, while already unlocked (by
//       different clients to get an auth token), we could be more explicit
//       about this.
func (s *service) unlock(ctx context.Context, req *AuthUnlockRequest) (string, error) {
	s.unlockMtx.Lock()
	defer s.unlockMtx.Unlock()

	// Unlock auth/vault (get token)
	token, err := s.auth.unlock(ctx, s.vault, req)
	if err != nil {
		return "", err
	}

	isNew := false

	// DB
	if !s.db.IsOpen() {
		logger.Infof("Opening %s...", cdbPath)
		path, err := s.env.AppPath(cdbPath, true)
		if err != nil {
			return "", err
		}

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
		dbk := keys.Bytes32(keys.HKDFSHA256(mk[:], 32, nil, []byte("keys.pub/cache")))
		if err := s.db.OpenAtPath(ctx, path, dbk); err != nil {
			return "", err
		}
	}

	// If database is new, we are either in a new state or from a uninstalled
	// (or migrated) state. In the uninstalled state, we should try to update
	// local db for any keys we have in our vault.
	if isNew {
		if err := s.updateAllKeys(ctx); err != nil {
			logger.Errorf("Failed to update keys on new database: %+v", err)
		}
	}

	s.startCheck()

	s.unlocked = true

	return token, nil
}

func (s *service) lock() {
	s.unlockMtx.Lock()
	defer s.unlockMtx.Unlock()

	if !s.unlocked {
		logger.Infof("Service already locked")
		return
	}

	s.stopCheck()

	s.auth.lock(s.vault)

	logger.Infof("Closing sdb...")
	s.db.Close()
	s.unlocked = false
}

func (s *service) tryCheck(ctx context.Context) {
	s.checkMtx.Lock()
	defer s.checkMtx.Unlock()

	logger.Debugf("Checking...")
	if _, err := s.vault.CheckSync(ctx, time.Duration(time.Minute)); err != nil {
		logger.Warningf("Failed to check sync: %v", err)
	}
	if ctx.Err() != nil {
		return
	}
	if err := s.checkKeys(ctx); err != nil {
		logger.Warningf("Failed to check keys: %v", err)
	}
}
