package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
)

// VaultSync (RPC) ...
func (s *service) VaultSync(ctx context.Context, req *VaultSyncRequest) (*VaultSyncResponse, error) {
	if err := s.vault.Sync(ctx); err != nil {
		return nil, errors.Wrapf(err, "failed to sync")
	}
	if err := s.checkForKeyUpdates(ctx); err != nil {
		return nil, errors.Wrapf(err, "failed to check for key updates")
	}
	return &VaultSyncResponse{}, nil
}

// VaultAuth (RPC) creates an auth phrase for the vault.
func (s *service) VaultAuth(ctx context.Context, req *VaultAuthRequest) (*VaultAuthResponse, error) {
	remote := s.vault.Remote()
	if remote == nil {
		return nil, errors.Errorf("no remote set")
	}

	// We don't want to export the remote vault key directly.
	// We create a temporary key, encrypt the auth and upload via the share API
	// with an expiry.

	b, err := json.Marshal(remote)
	if err != nil {
		return nil, err
	}
	otk := keys.GenerateEdX25519Key()
	if err := s.client.ShareSeal(ctx, otk, b[:], time.Minute*5); err != nil {
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

// VaultUpdate (RPC) ...
func (s *service) VaultUpdate(ctx context.Context, req *VaultUpdateRequest) (*VaultUpdateResponse, error) {
	// TODO: Add test to ensure syncing isn't accidentally activated
	if err := s.vaultUpdate(ctx, time.Duration(0)); err != nil {
		return nil, err
	}
	return &VaultUpdateResponse{}, nil
}

func (s *service) vaultUpdate(ctx context.Context, expire time.Duration) error {
	synced, err := s.vault.CheckSync(ctx, expire)
	if err != nil {
		return err
	}
	if synced {
		return s.checkForKeyUpdates(ctx)
	}
	return nil
}

// VaultStatus (RPC) ...
func (s *service) VaultStatus(ctx context.Context, req *VaultStatusRequest) (*VaultStatusResponse, error) {
	status, err := s.vault.SyncStatus()
	if err != nil {
		return nil, err
	}
	if status == nil {
		return &VaultStatusResponse{}, nil
	}
	return &VaultStatusResponse{
		KID:      status.KID.String(),
		SyncedAt: tsutil.Millis(status.SyncedAt),
	}, nil
}
