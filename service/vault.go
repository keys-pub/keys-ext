package service

import (
	"context"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/encoding"
	"github.com/pkg/errors"
)

// VaultSync (RPC) ...
func (s *service) VaultSync(ctx context.Context, req *VaultSyncRequest) (*VaultSyncResponse, error) {
	if err := s.vault.Sync(ctx); err != nil {
		return nil, err
	}
	if err := s.checkForKeyUpdates(ctx); err != nil {
		return nil, err
	}
	return &VaultSyncResponse{}, nil
}

// VaultAuth (RPC) creates an auth phrase for the vault.
func (s *service) VaultAuth(ctx context.Context, req *VaultAuthRequest) (*VaultAuthResponse, error) {
	rk := s.vault.RemoteKey()
	if rk == nil {
		return nil, errors.Errorf("no vault key set")
	}

	// We don't want to export the remote vault key seed bytes directly.
	// We create a key, encrypt the remote vault key seed bytes and upload via
	// the share remote API with an expiry.

	b := rk.Seed()
	otk := keys.GenerateEdX25519Key()
	if err := s.remote.ShareSeal(ctx, otk, b[:], time.Minute*5); err != nil {
		return nil, err
	}

	// Return BIP39 encoded bytes for one time key (otk) seed.
	out, err := encoding.Encode(otk.Seed()[:], encoding.BIP39)
	if err != nil {
		return nil, err
	}

	return &VaultAuthResponse{
		Phrase: out,
	}, nil
}

// VaultUnsync (RPC) ...
func (s *service) VaultUnsync(ctx context.Context, req *VaultUnsyncRequest) (*VaultUnsyncResponse, error) {
	if err := s.vault.Unsync(ctx); err != nil {
		return nil, err
	}
	return &VaultUnsyncResponse{}, nil
}
