package service

import (
	"context"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/git"
	"github.com/keys-pub/keys/keyring"
	"github.com/pkg/errors"
)

type gitKeyringFn struct {
	sync.Mutex
	git  *keyring.Keyring
	repo *git.Repository
}

func newGitKeyringFn(cfg *Config) (*gitKeyringFn, error) {
	path, err := gitPath(cfg)
	if err != nil {
		return nil, err
	}
	if path == "" {
		return nil, errors.Errorf("no git path specified")
	}
	repo := git.NewRepository()
	if err := repo.Open(path); err != nil {
		return nil, errors.Wrapf(err, "failed to open git repo")
	}

	git, err := keyring.New(keyring.WithStore(repo))
	if err != nil {
		return nil, err
	}

	key, err := loadGitKey(cfg)
	if err != nil {
		return nil, err
	}
	if err := repo.SetKey(key); err != nil {
		return nil, err
	}

	gitFn := &gitKeyringFn{
		git:  git,
		repo: repo,
	}
	return gitFn, nil
}

func startGitKeyringFn(cfg *Config) (*gitKeyringFn, error) {
	kr, err := newGitKeyringFn(cfg)
	if err != nil {
		return nil, err
	}
	kr.start()
	return kr, nil
}

func (k *gitKeyringFn) start() {
	go func() {
		// Initial sync
		k.sync()

		ch := k.git.Subscribe("git")
		timer := time.NewTimer(15 * time.Minute)
		for {
			select {
			// TODO: Implement stop, using context cancel
			// case <-ctx.Done():
			case <-timer.C:
				go func() {
					k.sync()
				}()
			case event := <-ch:
				switch event.(type) {
				case keyring.CreateEvent, keyring.UpdateEvent:
					logger.Infof("Create update event: %v", event)
					go func() {
						k.sync()
					}()
				}
			}
		}
	}()
}

func (k *gitKeyringFn) sync() {
	k.Lock()
	defer k.Unlock()

	// TODO: Don't add/remove items from keyring while sync'ing?

	if err := k.repo.Sync(); err != nil {
		logger.Errorf("Failed to sync git keyring: %v", err)
	}
}

func (k *gitKeyringFn) Keyring() *keyring.Keyring {
	return k.git
}

func gitPath(cfg *Config) (string, error) {
	path, err := cfg.keyringGitPath()
	if err != nil {
		return "", err
	}
	exists, err := pathExists(path)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", nil
	}
	return path, nil
}

// GitImport (RPC) imports into a git keyring.
func (s *service) GitImport(ctx context.Context, req *GitImportRequest) (*GitImportResponse, error) {
	urs := req.URL
	key, err := gitKey(req.KeyPath)
	if err != nil {
		return nil, err
	}
	if err := saveGitKeyPath(s.cfg, req.KeyPath); err != nil {
		return nil, err
	}
	if err := s.gitImport(ctx, key, urs); err != nil {
		return nil, err
	}
	return &GitImportResponse{}, nil
}

func (s *service) checkGitSetup() (string, error) {
	path, err := s.cfg.keyringGitPath()
	if err != nil {
		return "", err
	}
	logger.Infof("Checking path: %s", path)
	exists, err := pathExists(path)
	if err != nil {
		return "", err
	}
	if exists {
		return "", errors.Errorf("git repository already exists")
	}

	// Check current keyring (not already git)
	kr := s.keyring()
	logger.Infof("Current store: %s", kr.Store().Name())
	if kr.Store().Name() == "git" {
		return "", errors.Errorf("git already set as keyring")
	}
	return path, nil
}

func (s *service) gitImport(ctx context.Context, key *keys.EdX25519Key, urs string) error {
	logger.Infof("Git import...")

	path, err := s.checkGitSetup()
	if err != nil {
		return err
	}

	logger.Infof("Git using key: %s", key.ID())

	// Clear tmp path (if it exists)
	tmpPath := path + ".tmp"
	tmpExists, err := pathExists(tmpPath)
	if err != nil {
		return err
	}
	if tmpExists {
		logger.Infof("Remove existing temp: %s", tmpPath)
		if err := os.RemoveAll(tmpPath); err != nil {
			return err
		}
	}
	defer func() { _ = os.RemoveAll(tmpPath) }()

	// Clone repo (into tmpPath)
	repo := git.NewRepository()
	if err := repo.SetKey(key); err != nil {
		return err
	}
	logger.Infof("Cloning repo: %s", urs)
	if err := repo.Clone(urs, tmpPath); err != nil {
		return errors.Wrapf(err, "failed to clone git repo")
	}

	// Check existing state of the repo is empty
	existing, err := repo.IDs(keyring.Reserved(), keyring.Hidden())
	if err != nil {
		return err
	}
	if !equalStrings(existing, []string{".git"}) {
		return errors.Errorf("import only supported on empty repository")
	}

	// Copy old keyring into git repo (still in tmp)
	logger.Infof("Copying keyring into git...")
	kr := s.keyring()
	ids, err := keyring.Copy(kr.Store(), repo)
	if err != nil {
		return err
	}
	logger.Infof("Keyring copied into git: %s", ids)

	logger.Infof("Pushing git keyring...")
	if err := repo.Push(); err != nil {
		return err
	}

	// Move repo into place (from tmpPath)
	logger.Infof("Moving into place: %s", path)
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}

	// Set git as the service keyring
	logger.Infof("Setting keyring to git...")
	git, err := startGitKeyringFn(s.cfg)
	if err != nil {
		return err
	}
	s.keyringFn = git

	// Unlock the new git keyring (if unlocked, otherwise is noop)
	git.Keyring().SetMasterKey(kr.MasterKey())

	// TODO: Test the new keyring before reseting old?

	// Backup and reset old keyring
	if _, err := s.backup(kr.Store()); err != nil {
		return err
	}
	logger.Infof("Resetting old keyring...")
	if err := kr.Reset(); err != nil {
		return err
	}

	logger.Infof("Git setup finished")

	return nil
}

// GitClone (RPC) clones a repository.
func (s *service) GitClone(ctx context.Context, req *GitCloneRequest) (*GitCloneResponse, error) {
	urs := req.URL
	key, err := gitKey(req.KeyPath)
	if err != nil {
		return nil, err
	}
	if err := saveGitKeyPath(s.cfg, req.KeyPath); err != nil {
		return nil, err
	}
	if err := s.gitClone(ctx, key, urs); err != nil {
		return nil, err
	}
	return &GitCloneResponse{}, nil
}

func gitKey(path string) (*keys.EdX25519Key, error) {
	b, err := ioutil.ReadFile(path) // #nosec
	if err != nil {
		return nil, err
	}

	skey, err := keys.ParseSSHKey(b, nil, true)
	if err != nil {
		return nil, err
	}
	key, ok := skey.(*keys.EdX25519Key)
	if !ok {
		return nil, errors.Errorf("unsupported key type for git clone, ed25519 expected")
	}

	return key, nil
}

func saveGitKeyPath(cfg *Config, path string) error {
	logger.Infof("Saving git auth to config: %s", path)
	cfg.Set(gitKeyPathCfgKey, path)
	if err := cfg.Save(); err != nil {
		return err
	}
	return nil
}

func loadGitKey(cfg *Config) (*keys.EdX25519Key, error) {
	path := cfg.Get(gitKeyPathCfgKey, homePath(".ssh", "id_ed25519"))
	if path == "" {
		return nil, errors.Errorf("no git key set in config")
	}
	return gitKey(path)
}

func (s *service) gitClone(ctx context.Context, key *keys.EdX25519Key, urs string) error {
	logger.Infof("Git clone...")

	path, err := s.checkGitSetup()
	if err != nil {
		return err
	}
	logger.Infof("Git using key: %s", key.ID())

	// Clone repo
	repo := git.NewRepository()
	if err := repo.SetKey(key); err != nil {
		return err
	}
	logger.Infof("Cloning repo: %s", urs)
	if err := repo.Clone(urs, path); err != nil {
		return errors.Wrapf(err, "failed to clone git repo")
	}

	kr := s.keyringFn.Keyring()

	// Set git as the service keyring
	logger.Infof("Setting keyring to git...")
	git, err := startGitKeyringFn(s.cfg)
	if err != nil {
		return err
	}
	s.keyringFn = git

	// Unlock the new git keyring (if set)
	git.Keyring().SetMasterKey(kr.MasterKey())

	logger.Infof("Git clone finished")
	return nil
}

func equalStrings(s1 []string, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := 0; i < len(s1); i++ {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}
