package service

import (
	"context"
	"os"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/git"
	"github.com/keys-pub/keys/keyring"
	"github.com/pkg/errors"
)

type gitKeyringFn struct {
	git  *keyring.Keyring
	repo *git.Repository
	kid  keys.ID
}

func newGitKeyringFn(cfg *Config) (KeyringFn, error) {
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

	service := cfg.keyringService()
	git, err := keyring.New(service, repo)
	if err != nil {
		return nil, err
	}

	gitAuth := cfg.GitAuth()
	if gitAuth == "" {
		return nil, errors.Errorf("no git auth set")
	}
	kid, err := keys.ParseID(gitAuth)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse git auth")
	}

	gitFn := &gitKeyringFn{
		git:  git,
		repo: repo,
		kid:  kid,
	}
	git.AddListener(gitFn)
	return gitFn, nil
}

func (k *gitKeyringFn) Keyring() *keyring.Keyring {
	return k.git
}

func (k *gitKeyringFn) Pull() error {
	return k.repo.Pull()
}

func (k *gitKeyringFn) Push() error {
	return k.repo.Push()
}

func (k *gitKeyringFn) Locked() {
	if err := k.repo.SetKey(nil); err != nil {
		logger.Errorf("Failed to clear git repo key on lock: %v", err)
	}
}

func (k *gitKeyringFn) Unlocked(p *keyring.Provision) {
	if err := k.loadAuth(); err != nil {
		logger.Errorf("Failed to set git repo key on unlock: %v", err)
	}
}

func (k *gitKeyringFn) loadAuth() error {
	// Set repo auth using key from git keyring
	ks := keys.NewStore(k.git)
	key, err := ks.EdX25519Key(k.kid)
	if err != nil {
		return err
	}
	if key == nil {
		return keys.NewErrNotFound(k.kid.String())
	}
	if err := k.repo.SetKey(key); err != nil {
		return err
	}
	return nil
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

// GitSetup (RPC) sets up git keyring.
func (s *service) GitSetup(ctx context.Context, req *GitSetupRequest) (*GitSetupResponse, error) {
	// Check if already setup
	path, err := s.cfg.keyringGitPath()
	if err != nil {
		return nil, err
	}
	exists, err := pathExists(path)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.Errorf("git repository already exists")
	}

	// Check current keyring (not already git)
	kro := s.keyring()
	kso := keys.NewStore(kro)
	if kro.Store().Name() == "git" {
		return nil, errors.Errorf("git already set as keyring")
	}

	// Get key from current keyring
	kid, err := keys.ParseID(req.KID)
	if err != nil {
		return nil, err
	}
	key, err := kso.EdX25519Key(kid)
	if err != nil {
		return nil, err
	}

	// Clear tmp path (if it exists)
	tmpPath := path + ".tmp"
	tmpExists, err := pathExists(tmpPath)
	if err != nil {
		return nil, err
	}
	if tmpExists {
		if err := os.RemoveAll(tmpPath); err != nil {
			return nil, err
		}
	}
	defer func() { _ = os.RemoveAll(tmpPath) }()

	// Clone repo (into tmpPath)
	repo := git.NewRepository()
	if err := repo.SetKey(key); err != nil {
		return nil, err
	}
	if err := repo.Clone(req.URL, tmpPath); err != nil {
		return nil, errors.Wrapf(err, "failed to clone git repo")
	}

	// Copy old keyring into git repo (still in tmp)
	serviceName := s.cfg.keyringService()
	krg, err := keyring.New(serviceName, repo)
	if err != nil {
		return nil, err
	}
	ids, err := keyring.Copy(kro, krg)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Keyring copied: %s", ids)

	// Save KID as git auth to config
	s.cfg.Set(gitAuthCfgKey, key.ID().String())
	if err := s.cfg.Save(); err != nil {
		return nil, err
	}

	// Move repo into place (from tmpPath)
	if err := os.Rename(tmpPath, path); err != nil {
		return nil, err
	}

	// Set git as the service keyring
	git, err := newGitKeyringFn(s.cfg)
	if err != nil {
		return nil, err
	}
	s.keyringFn = git

	if err := repo.Push(); err != nil {
		return nil, err
	}

	// TODO: Test the new keyring before reseting old?

	// Reset old keyring
	if err := kro.Reset(); err != nil {
		return nil, err
	}

	return &GitSetupResponse{}, nil
}
