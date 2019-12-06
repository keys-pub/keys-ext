package service

import (
	"context"
	"os"

	"github.com/pkg/errors"
)

// Status (RPC) returns status.
func (s *service) Status(ctx context.Context, req *StatusRequest) (*StatusResponse, error) {
	logger.Infof("Status")
	key, err := s.loadCurrentKey()
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, errors.Errorf("no current key")
	}

	keyOut, err := s.key(ctx, key.ID(), true, false)
	if err != nil {
		return nil, err
	}

	url := ""
	if s.remote != nil {
		url = s.remote.URL().String()
	}

	promptPublish := keyOut.PublishedAt == 0 && !s.cfg.DisablePromptPublish()
	promptUser := len(keyOut.Users) == 0 && !s.cfg.DisablePromptUser()

	return &StatusResponse{
		URI:           url,
		Key:           keyOut,
		PromptPublish: promptPublish,
		PromptUser:    promptUser,
	}, nil
}

// RuntimeStatus (RPC) gets the current runtime status.
// This call is NOT AUTHENTICATED.
func (s *service) RuntimeStatus(ctx context.Context, req *RuntimeStatusRequest) (*RuntimeStatusResponse, error) {
	runtime := os.Getenv("KEYS_RUNTIME")
	label := os.Getenv("KEYS_LABEL")
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
		Runtime:         runtime,
		Label:           label,
		Exe:             exe,
		AuthSetupNeeded: !authed,
	}
	logger.Infof("Runtime status, %s", resp.String())
	return &resp, nil
}
