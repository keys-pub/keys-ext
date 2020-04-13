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
	ks     *keys.Keystore
	ss     *secret.Store
	remote *client.Client
	scs    keys.SigchainStore
	users  *user.Store
	nowFn  func() time.Time

	closeCh chan bool
	open    bool
	openMtx sync.Mutex

	watchLast *ds.WatchEvent
	watchLn   ds.WatchLn
	watchWg   *sync.WaitGroup
	watchMtx  sync.Mutex
}

func newService(cfg *Config, build Build, auth *auth, req keys.Requestor, nowFn func() time.Time) (*service, error) {
	ks := keys.NewKeystore(auth.keyring)
	ss := secret.NewStore(auth.keyring)
	ss.SetTimeNow(nowFn)
	db := db.NewDB()
	db.SetTimeNow(nowFn)
	scs := keys.NewSigchainStore(db)
	users, err := user.NewStore(db, scs, req, nowFn)
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
	path, err := s.cfg.AppPath("keys-v2.leveldb", true)
	if err != nil {
		return err
	}

	isNew := false
	if _, err := os.Stat(path); os.IsNotExist(err) {
		isNew = true
	}

	if err := s.db.OpenAtPath(ctx, path, key); err != nil {
		return err
	}
	s.open = true

	// If database is new, we are either in a new state or from a uninstalled
	// (or migrated) state. In the uninstalled state, we should try to update
	// local db for any keys we have in our keyring.
	if isNew {
		s.tryCheckUpdate(ctx)
	}

	s.startCheck()

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

func (s *service) tryCheckUpdate(ctx context.Context) {
	if err := s.checkUpdate(ctx); err != nil {
		logger.Errorf("Failed to check keys: %v", err)
	}
}

func (s *service) checkUpdate(ctx context.Context) error {
	logger.Infof("Checking keys...")

	// TODO: Only update keys where we've seen a sigchain?

	pks, err := s.ks.EdX25519PublicKeys()
	if err != nil {
		return err
	}
	kids := make([]keys.ID, 0, len(pks))
	for _, pk := range pks {
		res, err := s.users.Get(ctx, pk.ID())
		if err != nil {
			return err
		}
		if res == nil || res.Expired(s.Now(), time.Hour*24) {
			kids = append(kids, pk.ID())
		}
	}

	for _, kid := range kids {
		if _, _, err := s.update(ctx, kid); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) startCheck() {
	ticker := time.NewTicker(time.Hour)
	s.closeCh = make(chan bool)
	go func() {
		s.tryCheckUpdate(context.TODO())
		for {
			select {
			case <-ticker.C:
				s.tryCheckUpdate(context.TODO())
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
