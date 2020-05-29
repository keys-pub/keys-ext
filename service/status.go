package service

import (
	"context"

	"github.com/keys-pub/keys/keyring"
)

// RuntimeStatus (RPC) gets the current runtime status.
// This call is NOT AUTHENTICATED.
func (s *service) RuntimeStatus(ctx context.Context, req *RuntimeStatusRequest) (*RuntimeStatusResponse, error) {
	exe, exeErr := executablePath()
	if exeErr != nil {
		logger.Errorf("Failed to get current executable path: %s", exeErr)
	}
	kr := s.auth.Keyring()
	status, err := kr.Status()
	if err != nil {
		return nil, err
	}

	resp := RuntimeStatusResponse{
		Version:    s.build.Version,
		AppName:    s.cfg.AppName(),
		Exe:        exe,
		AuthStatus: keyringStatusToRPC(status),
		FIDO2:      s.auth.auths != nil,
	}
	logger.Infof("Runtime status, %s", resp.String())
	return &resp, nil
}

func keyringStatusToRPC(st keyring.Status) AuthStatus {
	switch st {
	case keyring.Locked:
		return AuthLocked
	case keyring.Unlocked:
		return AuthUnlocked
	case keyring.Setup:
		return AuthSetup
	default:
		return AuthUnknown
	}
}
