package service

import (
	"context"

	"github.com/keys-pub/keys"
)

// RuntimeStatus (RPC) gets the current runtime status.
// This call is NOT AUTHENTICATED.
func (s *service) RuntimeStatus(ctx context.Context, req *RuntimeStatusRequest) (*RuntimeStatusResponse, error) {
	exe, exeErr := ExecutablePath()
	if exeErr != nil {
		logger.Errorf("Failed to get current executable path: %s", exeErr)
	}
	kr, err := s.ks.Keyring()
	if err != nil {
		return nil, err
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

func (s *service) AppStatus(ctx context.Context, req *AppStatusRequest) (*AppStatusResponse, error) {
	ks, err := s.ks.Keys(&keys.Opts{
		Types: []keys.KeyType{keys.EdX25519, keys.X25519},
	})
	if err != nil {
		return nil, err
	}
	hasUser := false
	for _, k := range ks {
		res, err := s.users.Get(ctx, k.ID())
		if err != nil {
			return nil, err
		}
		if res != nil {
			hasUser = true
			break
		}
	}

	promptKeygen := len(ks) == 0 && !s.cfg.DisablePromptKeygen()

	promtUser := !hasUser && !s.cfg.DisablePromptUser()

	return &AppStatusResponse{
		PromptKeygen: promptKeygen,
		PromptUser:   promtUser,
	}, nil
}
