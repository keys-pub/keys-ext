package service

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// Pull (RPC)
func (s *service) Pull(ctx context.Context, req *PullRequest) (*PullResponse, error) {
	if req.Identity != "" {
		kid, err := s.searchIdentity(context.TODO(), req.Identity)
		if err != nil {
			return nil, err
		}
		ok, err := s.pull(ctx, kid)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, keys.NewErrNotFound(kid.String())
		}
		return &PullResponse{KIDs: []string{kid.String()}}, nil
	}

	// Update existing if no kid or user specified
	pulled := []string{}
	spks, err := s.ks.SignPublicKeys()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load keys")
	}
	for _, spk := range spks {
		ok, err := s.pull(ctx, spk.ID())
		if err != nil {
			return nil, err
		}
		if !ok {
			// TODO: Report missing
			continue
		}
		pulled = append(pulled, spk.ID().String())
	}
	return &PullResponse{KIDs: pulled}, nil
}

func (s *service) pull(ctx context.Context, kid keys.ID) (bool, error) {
	logger.Infof("Pull %s", kid)

	if err := s.importID(kid); err != nil {
		return false, err
	}

	return s.update(ctx, kid)
}

func (s *service) update(ctx context.Context, kid keys.ID) (bool, error) {
	logger.Infof("Update %s", kid)

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
