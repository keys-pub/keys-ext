package service

import (
	"context"
	"fmt"
	"time"

	"github.com/keys-pub/keys/keyring"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
)

func (s *service) backup(st keyring.Store) (string, error) {
	now := time.Now()
	backupFile := fmt.Sprintf("keyring-backup-%d.tgz", tsutil.Millis(now))
	backupPath, err := s.cfg.AppPath(backupFile, true)
	if err != nil {
		return "", err
	}
	logger.Infof("Backing up to %s", backupPath)
	if err := keyring.Backup(backupPath, st, now); err != nil {
		return "", err
	}
	return backupPath, nil
}

// Backup (RPC) ...
func (s *service) Backup(ctx context.Context, req *BackupRequest) (*BackupResponse, error) {
	st := s.keyringFn.Keyring().Store()
	if st.Name() == "git" {
		return nil, errors.Errorf("keyring backup not supported for git repo")
	}

	path, err := s.backup(st)
	if err != nil {
		return nil, err
	}
	return &BackupResponse{
		Path: path,
	}, nil
}

// Restore (RPC) ...
func (s *service) Restore(ctx context.Context, req *RestoreRequest) (*RestoreResponse, error) {
	st := s.keyringFn.Keyring().Store()
	if st.Name() == "git" {
		return nil, errors.Errorf("keyring restore not supported for git repo")
	}

	if err := keyring.Restore(req.Path, st); err != nil {
		return nil, err
	}
	return &RestoreResponse{}, nil
}
