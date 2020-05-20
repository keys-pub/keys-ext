package service

import (
	"context"
	"os"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/git"
	"github.com/pkg/errors"
)

// GitSetup (RPC) sets up git keyring.
func (s *service) GitSetup(ctx context.Context, req *GitSetupRequest) (*GitSetupResponse, error) {
	path, err := s.cfg.AppPath("keyring", true)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); err == nil {
		return nil, errors.Errorf("git keyring already setup")
	} else if os.IsNotExist(err) {
		// OK
	} else {
		return nil, err
	}

	key, err := keys.ParseSSHKey([]byte(req.Key), nil, true)
	if err != nil {
		return nil, err
	}

	repo, err := git.NewRepository(req.URL, path, key, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to setup new git repo")
	}
	if err := repo.Open(); err != nil {
		return nil, errors.Wrapf(err, "failed to open git repo")
	}

	// service := s.cfg.AppName()

	// kr, err := keyring.New(service, repo)
	// if err != nil {
	// 	return nil, err
	// }

	return &GitSetupResponse{}, nil
}
