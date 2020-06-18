package service

import (
	"context"

	"github.com/keys-pub/keys-ext/vault"
)

// RuntimeStatus (RPC) gets the current runtime status.
// This call is NOT AUTHENTICATED.
func (s *service) RuntimeStatus(ctx context.Context, req *RuntimeStatusRequest) (*RuntimeStatusResponse, error) {
	exe, exeErr := executablePath()
	if exeErr != nil {
		logger.Errorf("Failed to get current executable path: %s", exeErr)
	}
	status, err := s.vault.Status()
	if err != nil {
		return nil, err
	}

	resp := RuntimeStatusResponse{
		Version:    s.build.Version,
		AppName:    s.cfg.AppName(),
		Exe:        exe,
		AuthStatus: vaultStatusToRPC(status),
		FIDO2:      s.auth.fas != nil,
	}
	logger.Infof("Runtime status, %s", resp.String())
	return &resp, nil
}

func vaultStatusToRPC(st vault.Status) AuthStatus {
	switch st {
	case vault.Locked:
		return AuthLocked
	case vault.Unlocked:
		return AuthUnlocked
	case vault.Setup:
		return AuthSetup
	default:
		return AuthUnknown
	}
}
