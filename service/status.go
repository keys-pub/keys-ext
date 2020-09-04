package service

import (
	"context"
	"time"

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

	sync, err := s.vault.SyncEnabled()
	if err != nil {
		return nil, err
	}

	// Check vault sync if unlocked
	if status == vault.Unlocked {
		go func() {
			if err := s.vaultUpdate(context.TODO(), time.Minute*5); err != nil {
				logger.Errorf("Failed to check sync: %v", err)
			}
		}()
	}

	resp := RuntimeStatusResponse{
		Version:    s.build.Version,
		AppName:    s.env.AppName(),
		Exe:        exe,
		AuthStatus: vaultStatusToRPC(status),
		Sync:       sync,
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

func (s *service) ensureUnlocked() error {
	status, err := s.vault.Status()
	if err != nil {
		return err
	}
	if status != vault.Unlocked {
		return vault.ErrLocked
	}
	return nil
}
