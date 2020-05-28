package service

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/git"
	"github.com/pkg/errors"
)

// GitSetup (RPC) sets up git keyring.
func (s *service) GitSetup(ctx context.Context, req *GitSetupRequest) (*GitSetupResponse, error) {
	path, err := s.cfg.keyringGitPath()
	if err != nil {
		return nil, err
	}
	exists, err := pathExists(path)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.Errorf("git keyring already setup")
	}

	key, err := keys.ParseSSHKey([]byte(req.Key), nil, true)
	if err != nil {
		return nil, err
	}

	repo := git.NewRepository()

	if err := repo.SetKey(key); err != nil {
		return nil, err
	}

	if err := repo.Clone(req.URL, path); err != nil {
		return nil, errors.Wrapf(err, "failed to clone git repo")
	}

	return &GitSetupResponse{}, nil
}
