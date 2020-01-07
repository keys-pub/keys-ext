package service

import (
	"context"

	"github.com/pkg/errors"
)

// Status (RPC) returns status.
func (s *service) Status(ctx context.Context, req *StatusRequest) (*StatusResponse, error) {
	logger.Infof("Status")

	var key *Key
	promptPublish := false
	promptUser := false

	sk, err := s.loadCurrentKey()
	if err != nil {
		return nil, err
	}
	if sk != nil {
		k, err := s.key(ctx, sk.ID())
		if err != nil {
			return nil, err
		}
		key = k
		promptPublish = key.PublishedAt == 0 && !s.cfg.DisablePromptPublish()
		promptUser = len(key.Users) == 0 && !s.cfg.DisablePromptUser()
	}

	url := ""
	if s.remote != nil {
		url = s.remote.URL().String()
	}

	return &StatusResponse{
		URI:           url,
		Key:           key,
		PromptPublish: promptPublish,
		PromptUser:    promptUser,
	}, nil
}

// RuntimeStatus (RPC) gets the current runtime status.
// This call is NOT AUTHENTICATED.
func (s *service) RuntimeStatus(ctx context.Context, req *RuntimeStatusRequest) (*RuntimeStatusResponse, error) {
	exe, exeErr := ExecutablePath()
	if exeErr != nil {
		logger.Errorf("Failed to get current executable path: %s", exeErr)
	}
	kr := s.ks.Keyring()
	if kr == nil {
		return nil, errors.Errorf("no keyring set")
	}
	authed, authedErr := kr.Authed()
	if authedErr != nil {
		return nil, authedErr
	}
	resp := RuntimeStatusResponse{
		Version:         s.build.Version,
		Exe:             exe,
		AuthSetupNeeded: !authed,
	}
	logger.Infof("Runtime status, %s", resp.String())
	return &resp, nil
}
