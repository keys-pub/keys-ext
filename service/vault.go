package service

import "context"

// VaultSync (RPC) ...
func (s *service) VaultSync(ctx context.Context, req *VaultSyncRequest) (*VaultSyncResponse, error) {
	if err := s.vault.Sync(ctx); err != nil {
		return nil, err
	}
	return &VaultSyncResponse{}, nil
}
