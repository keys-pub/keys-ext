package service

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// Pull (RPC)
func (s *service) Pull(ctx context.Context, req *PullRequest) (*PullResponse, error) {
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

	// Update users
	if _, err = s.users.Update(ctx, kid); err != nil {
		return false, err
	}

	return true, nil
}
