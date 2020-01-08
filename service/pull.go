package service

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// Pull (RPC)
func (s *service) Pull(ctx context.Context, req *PullRequest) (*PullResponse, error) {
	if req.All {
		if req.KID != "" || req.User != "" {
			return nil, errors.Errorf("all specified with other arguments")
		}

		kids, err := s.pullStatements(ctx)
		if err != nil {
			return nil, err
		}
		return &PullResponse{KIDs: kids}, nil
	}

	if req.KID != "" {
		kid, err := keys.ParseID(req.KID)
		if err != nil {
			return nil, err
		}
		ok, err := s.pull(ctx, kid)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, keys.NewErrNotFound(req.KID)
		}
		return &PullResponse{KIDs: []string{kid.String()}}, nil
	} else if req.User != "" {
		usr, err := s.searchUserExact(ctx, req.User)
		if err != nil {
			return nil, err
		}
		if usr == nil {
			return &PullResponse{}, nil
		}
		ok, err := s.pull(ctx, usr.User.KID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, errors.Errorf("%s not found", req.User)
		}
		return &PullResponse{KIDs: []string{usr.User.KID.String()}}, nil
	}

	// Update existing if no kid or user specified
	pulled := []string{}
	kids, err := s.loadKIDs(true)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load kids")
	}
	for _, kid := range kids {
		ok, err := s.pull(ctx, kid)
		if err != nil {
			return nil, err
		}
		if !ok {
			// TODO: Report missing
			continue
		}
		pulled = append(pulled, kid.String())
	}
	return &PullResponse{KIDs: pulled}, nil
}

func (s *service) pull(ctx context.Context, kid keys.ID) (bool, error) {
	if s.remote == nil {
		return false, errors.Errorf("no remote set")
	}
	if s.db == nil {
		return false, errors.Errorf("db is locked")
	}
	logger.Infof("Pull sigchain %s", kid)
	resp, err := s.remote.Sigchain(kid)
	if err != nil {
		return false, err
	}
	if resp == nil {
		logger.Infof("No sigchain for %s", kid)
		return false, nil
	}
	logger.Infof("Received sigchain %s, len=%d", kid, len(resp.Statements))
	for _, st := range resp.Statements {
		if err := s.db.Set(ctx, keys.Path("sigchain", st.Key()), st.Bytes()); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (s *service) pullStatements(ctx context.Context) ([]string, error) {
	logger.Infof("Pull statements...")
	versionPath := keys.Path("versions", "sigchains")
	e, err := s.db.Get(ctx, versionPath)
	if err != nil {
		return nil, err
	}

	version := ""
	if e != nil {
		version = string(e.Data)
	}
	if s.remote == nil {
		return nil, errors.Errorf("no remote set")
	}
	resp, err := s.remote.Sigchains(version)
	if err != nil {
		return nil, err
	}
	kids := keys.NewStringSet()
	for _, st := range resp.Statements {
		if err := s.db.Set(ctx, keys.Path("sigchain", st.Key()), st.Bytes()); err != nil {
			return nil, err
		}
		kids.Add(st.KID.String())
	}
	if err := s.db.Set(ctx, versionPath, []byte(resp.Version)); err != nil {
		return nil, err
	}
	return kids.Strings(), nil
}
