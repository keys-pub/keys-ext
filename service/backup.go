package service

import (
	"context"
	"fmt"
	"time"

	"github.com/keys-pub/keys/keyring"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
)

// Backup (RPC) ...
func (s *service) Backup(ctx context.Context, req *BackupRequest) (*BackupResponse, error) {
	switch req.Resource {
	case "keyring":
		st := s.kr.Store()
		path, err := backupKeyring(s.cfg, st)
		if err != nil {
			return nil, err
		}
		return &BackupResponse{
			Path: path,
		}, nil
	case "":
		return nil, errors.Errorf("no resource specified")
	default:
		return nil, errors.Errorf("unrecognized resource: %s", req.Resource)
	}

}

// Restore (RPC) ...
func (s *service) Restore(ctx context.Context, req *RestoreRequest) (*RestoreResponse, error) {
	switch req.Resource {
	case "keyring":
		st := s.kr.Store()
		if err := keyring.Restore(req.Path, st); err != nil {
			return nil, err
		}
		return &RestoreResponse{}, nil
	case "":
		return nil, errors.Errorf("no resource specified")
	default:
		return nil, errors.Errorf("unrecognized resource: %s", req.Resource)
	}
}

// Migrate (RPC) ...
func (s *service) Migrate(ctx context.Context, req *MigrateRequest) (*MigrateResponse, error) {
	switch req.Resource {
	case "keyring":
		if err := s.migrateKeyring(req.Source, req.Destination); err != nil {
			return nil, err
		}
		return &MigrateResponse{}, nil
	case "":
		return nil, errors.Errorf("no resource specified")
	default:
		return nil, errors.Errorf("unrecognized resource: %s", req.Resource)
	}
}

func (s *service) migrateKeyring(source string, destination string) error {
	// So we can unlock new keyring after
	mk := s.kr.MasterKey()

	if err := migrateKeyring(s.cfg, source, destination); err != nil {
		return err
	}

	// (Re-)load keyring
	kr, scfg, err := newKeyring(s.cfg, "")
	if err != nil {
		return err
	}
	kr.SetMasterKey(mk)
	s.kr = kr
	s.scfg = scfg

	return nil
}

func backupKeyring(cfg *Config, st keyring.Store) (string, error) {
	now := time.Now()
	backupFile := fmt.Sprintf("backup-keyring-%d.tgz", tsutil.Millis(now))
	backupPath, err := cfg.AppPath(backupFile, true)
	if err != nil {
		return "", err
	}
	logger.Infof("Backing up to %s", backupPath)
	if err := keyring.Backup(backupPath, st, now); err != nil {
		return "", err
	}
	return backupPath, nil
}
