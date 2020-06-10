package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/keys-pub/keys-ext/syncp"
	"github.com/keys-pub/keys/ds"
	"github.com/pkg/errors"
)

// Sync (RPC) ...
func (s *service) Sync(req *SyncRequest, srv Keys_SyncServer) error {
	ctx := srv.Context()
	name, location := strings.TrimSpace(req.Name), strings.TrimSpace(req.Location)
	rt := syncRuntime{srv: srv}
	if err := s.sync(ctx, name, location, syncp.WithRuntime(rt)); err != nil {
		return err
	}
	return nil
}

func (s *service) sync(ctx context.Context, name string, location string, opt ...syncp.SyncOption) error {
	// Set remote if name, location
	remoteSet := false
	if name != "" {
		set, err := s.syncRemoteSet(ctx, name, location)
		if err != nil {
			return err
		}
		remoteSet = set
	}

	// Get remote
	sr, err := s.syncRemote(ctx)
	if err != nil {
		return err
	}
	if sr == nil {
		return errors.Errorf("no sync remote is set")
	}

	// Sync
	program, err := syncp.NewProgram(sr.Name, sr.Location)
	if err != nil {
		return err
	}
	if err := program.Sync(s.syncConfig(), opt...); err != nil {
		// If program was set (new) in this call, we'll unset it.
		if remoteSet {
			if err := s.syncUnset(ctx); err != nil {
				logger.Errorf("Unable to unset remote after failure: %v", err)
			}
		}
		return err
	}
	return nil
}

type syncRuntime struct {
	srv Keys_SyncServer
}

func (s syncRuntime) Log(format string, args ...interface{}) {
	if err := s.srv.Send(&SyncOutput{
		Out: fmt.Sprintf(format, args...),
	}); err != nil {
		logger.Errorf("Failed to send sync output to client: %v", err)
	}
}

func (s *service) syncRemote(ctx context.Context) (*SyncRemote, error) {
	doc, err := s.db.Get(ctx, ds.Path("sync", "remote"))
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}
	var sr SyncRemote
	if err := json.Unmarshal(doc.Data, &sr); err != nil {
		return nil, err
	}
	return &sr, nil
}

// SyncSet (RPC) ...
func (s *service) SyncSet(ctx context.Context, req *SyncSetRequest) (*SyncSetResponse, error) {
	name, location := strings.TrimSpace(req.Name), strings.TrimSpace(req.Location)
	if name != "" {
		_, err := s.syncRemoteSet(ctx, name, location)
		if err != nil {
			return nil, err
		}
	}

	remote, err := s.syncRemote(ctx)
	if err != nil {
		return nil, err
	}
	return &SyncSetResponse{
		Remote: remote,
	}, nil
}

func (s *service) syncRemoteSet(ctx context.Context, name string, location string) (bool, error) {
	if err := s.syncSupported(); err != nil {
		return false, err
	}

	// Validate program
	_, err := syncp.NewProgram(name, location)
	if err != nil {
		return false, err
	}
	sr, err := s.syncRemote(ctx)
	if err != nil {
		return false, err
	}
	if sr == nil {
		// Nothing set yet
	} else if sr.Name == name && sr.Location == location {
		// Resetting to same, it is already set
		return false, nil
	} else {
		return false, errors.Errorf("sync remote is already set")
	}

	srNew := &SyncRemote{
		Name:     name,
		Location: location,
	}
	b, err := json.Marshal(srNew)
	if err != nil {
		return false, err
	}
	if err := s.db.Set(ctx, ds.Path("sync", "remote"), b); err != nil {
		return false, err
	}
	return true, nil
}

// SyncUnset (RPC) ...
func (s *service) SyncUnset(ctx context.Context, req *SyncUnsetRequest) (*SyncUnsetResponse, error) {
	if err := s.syncUnset(ctx); err != nil {
		return nil, err
	}
	return &SyncUnsetResponse{}, nil
}

func (s *service) syncUnset(ctx context.Context) error {
	sr, err := s.syncRemote(ctx)
	if err != nil {
		return err
	}
	if sr == nil {
		return errors.Errorf("no sync remote is set")
	}

	ok, err := s.db.Delete(ctx, ds.Path("sync", "remote"))
	if err != nil {
		return err
	}
	if !ok {
		return errors.Errorf("no sync remote is set")
	}

	program, err := syncp.NewProgram(sr.Name, sr.Location)
	if err != nil {
		return err
	}
	if err := program.Clean(s.syncConfig()); err != nil {
		return err
	}

	return nil
}

func (s *service) syncSupported() error {
	if s.syncConfig().Dir == "" {
		return errors.Errorf("sync not supported for current keyring type")
	}
	return nil
}

func (s *service) syncConfig() syncp.Config {
	return syncp.Config{}
}
