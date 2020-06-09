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

	programSet := false
	name, remote := strings.TrimSpace(req.Name), strings.TrimSpace(req.Remote)
	if name != "" {
		set, err := s.syncProgramSet(ctx, name, remote)
		if err != nil {
			return err
		}
		programSet = set
	}

	sp, err := s.syncProgram(ctx)
	if err != nil {
		return err
	}
	if sp == nil {
		return errors.Errorf("no sync program is set")
	}

	rt := syncRuntime{srv: srv}
	program, err := syncp.NewProgram(sp.Name, sp.Remote)
	if err != nil {
		return err
	}
	if err := program.Sync(s.scfg, syncp.WithRuntime(rt)); err != nil {
		// If program was set (new) in this call, we'll unset it.
		if programSet {
			if err := s.syncUnset(ctx); err != nil {
				logger.Errorf("Unable to unset program after failure: %v", err)
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

func syncProgram(p *SyncProgram) (syncp.Program, error) {
	return syncp.NewProgram(p.Name, p.Remote)
}

func (s *service) syncProgram(ctx context.Context) (*SyncProgram, error) {
	doc, err := s.db.Get(ctx, ds.Path("sync", "program"))
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}
	var sp SyncProgram
	if err := json.Unmarshal(doc.Data, &sp); err != nil {
		return nil, err
	}
	return &sp, nil
}

// SyncPrograms (RPC) ...
func (s *service) SyncProgram(ctx context.Context, req *SyncProgramRequest) (*SyncProgramResponse, error) {
	sp, err := s.syncProgram(ctx)
	if err != nil {
		return nil, err
	}
	return &SyncProgramResponse{
		Program: sp,
	}, nil
}

func (s *service) syncProgramSet(ctx context.Context, name string, remote string) (bool, error) {
	// Validate program
	_, err := syncp.NewProgram(name, remote)
	if err != nil {
		return false, err
	}
	sp, err := s.syncProgram(ctx)
	if err != nil {
		return false, err
	}
	if sp == nil {
		// Nothing set yet
	} else if sp.Name == name && sp.Remote == remote {
		// Resetting to same, it is already set
		return false, nil
	} else {
		return false, errors.Errorf("sync program is already set")
	}

	program := &SyncProgram{
		Name:   name,
		Remote: remote,
	}
	b, err := json.Marshal(program)
	if err != nil {
		return false, err
	}
	if err := s.db.Set(ctx, ds.Path("sync", "program"), b); err != nil {
		return false, err
	}
	return true, nil
}

func (s *service) syncUnset(ctx context.Context) error {
	sp, err := s.syncProgram(ctx)
	if err != nil {
		return err
	}
	if sp == nil {
		return errors.Errorf("no sync program is set")
	}

	ok, err := s.db.Delete(ctx, ds.Path("sync", "program"))
	if err != nil {
		return err
	}
	if !ok {
		return errors.Errorf("no sync program is set")
	}

	program, err := syncp.NewProgram(sp.Name, sp.Remote)
	if err != nil {
		return err
	}
	if err := program.Clean(s.scfg); err != nil {
		return err
	}

	return nil
}

// SyncUnset (RPC) ...
func (s *service) SyncUnset(ctx context.Context, req *SyncUnsetRequest) (*SyncUnsetResponse, error) {
	if err := s.syncUnset(ctx); err != nil {
		return nil, err
	}
	return &SyncUnsetResponse{}, nil
}
