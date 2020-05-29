package service

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/git"
	"github.com/keys-pub/keys/keyring"
	"github.com/pkg/errors"
)

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
	kro := s.auth.Keyring()
	if kro.Store().Name() == "git" {
		return nil, errors.Errorf("git already set as keyring")
	}

	// Clone repo
	repo := git.NewRepository()
	key, err := keys.ParseSSHKey([]byte(req.Key), nil, true)
	if err != nil {
		return nil, err
	}
	if err := repo.SetKey(key); err != nil {
		return nil, err
	}
	if err := repo.Clone(req.URL, path); err != nil {
		return nil, errors.Wrapf(err, "failed to clone git repo")
	}

	// New git keyring
	serviceName := s.cfg.keyringService()
	krg, err := keyring.New(serviceName, repo)
	if err != nil {
		// TODO: If we fail here we are are in an inconsistent state
		return nil, err
	}

	// Copy old keyring into git repo
	ids, err := keyring.Copy(kro, krg)
	if err != nil {
		// TODO: If we fail here we are are in an inconsistent state
		return nil, err
	}
	logger.Debugf("Keyring copied: %s", ids)

	// Set new git keyring store
	ok, err := s.auth.loadGit()
	if err != nil {
		// TODO: If we fail here we are are in an inconsistent state
		return nil, err
	}
	if !ok {
		// TODO: If we fail here we are are in an inconsistent state
		return nil, errors.Errorf("failed to load git after setup")
	}

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
