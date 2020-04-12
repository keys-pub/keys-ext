package service

import (
	"context"
)

// RuntimeStatus (RPC) gets the current runtime status.
// This call is NOT AUTHENTICATED.
func (s *service) RuntimeStatus(ctx context.Context, req *RuntimeStatusRequest) (*RuntimeStatusResponse, error) {
	exe, exeErr := executablePath()
	if exeErr != nil {
		logger.Errorf("Failed to get current executable path: %s", exeErr)
	}
	kr := s.ks.Keyring()
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
