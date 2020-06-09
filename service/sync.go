package service

import (
	"context"
	"encoding/json"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/syncp"
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/encoding"
	"github.com/pkg/errors"
)

// Sync (RPC) ...
func (s *service) Sync(ctx context.Context, req *SyncRequest) (*SyncResponse, error) {
	programs, err := s.syncPrograms(ctx)
	if err != nil {
		return nil, err
	}
	rt := syncp.NewRuntime()
	for _, p := range programs {
		program, err := syncProgram(p)
		if err != nil {
			return nil, err
		}
		if err := program.Sync(s.scfg, rt); err != nil {
			return nil, err
		}
	}
	return &SyncResponse{}, nil
}

func syncProgram(p *SyncProgram) (syncp.Program, error) {
	return syncp.NewProgram(p.Name, p.Remote)
}

func (s *service) syncPrograms(ctx context.Context) ([]*SyncProgram, error) {
	iter, err := s.db.Documents(ctx, ds.Path("/sync.programs"))
	if err != nil {
		return nil, err
	}
	programs := []*SyncProgram{}
	for {
		doc, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if doc == nil {
			break
		}
		var sp SyncProgram
		if err := json.Unmarshal(doc.Data, &sp); err != nil {
			return nil, err
		}
		programs = append(programs, &sp)
	}

	return programs, nil
}

// SyncPrograms (RPC) ...
func (s *service) SyncPrograms(ctx context.Context, req *SyncProgramsRequest) (*SyncProgramsResponse, error) {
	programs, err := s.syncPrograms(ctx)
	if err != nil {
		return nil, err
	}
	// TODO: Sort by priority
	return &SyncProgramsResponse{
		Programs: programs,
	}, nil
}

// SyncProgramsAdd (RPC) ...
func (s *service) SyncProgramsAdd(ctx context.Context, req *SyncProgramsAddRequest) (*SyncProgramsAddResponse, error) {
	// Validate program
	sp, err := syncp.NewProgram(req.Name, req.Remote)
	if err != nil {
		return nil, err
	}

	rt := syncp.NewRuntime()
	if err := sp.Setup(s.scfg, rt); err != nil {
		return nil, err
	}

	id := encoding.MustEncode(keys.RandBytes(32), encoding.Base62)
	program := &SyncProgram{
		ID:     id,
		Name:   req.Name,
		Remote: req.Remote,
	}
	b, err := json.Marshal(program)
	if err != nil {
		return nil, err
	}
	if err := s.db.Set(ctx, ds.Path("sync.programs", id), b); err != nil {
		return nil, err
	}
	return &SyncProgramsAddResponse{
		Program: program,
	}, nil
}

func (s *service) findProgram(ctx context.Context, id string) (*SyncProgram, error) {
	doc, err := s.db.Get(ctx, ds.Path("sync.programs", id))
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

// SyncProgramsRemove (RPC) ...
func (s *service) SyncProgramsRemove(ctx context.Context, req *SyncProgramsRemoveRequest) (*SyncProgramsRemoveResponse, error) {
	if req.ID == "" {
		return nil, errors.Errorf("no id specified")
	}

	program, err := s.findProgram(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if program == nil {
		return nil, errors.Errorf("program not found %s", req.ID)
	}

	ok, err := s.db.Delete(ctx, ds.Path("sync.programs", req.ID))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.Errorf("program not found %s", req.ID)
	}

	sp, err := syncp.NewProgram(program.Name, program.Remote)
	if err != nil {
		return nil, err
	}
	rt := syncp.NewRuntime()
	if err := sp.Clean(s.scfg, rt); err != nil {
		return nil, err
	}

	return &SyncProgramsRemoveResponse{}, nil
}
