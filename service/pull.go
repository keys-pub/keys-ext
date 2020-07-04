package service

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/docs"
	"github.com/keys-pub/keys/user"
	"github.com/pkg/errors"
)

// Pull (RPC)
func (s *service) Pull(ctx context.Context, req *PullRequest) (*PullResponse, error) {
	if req.Identity != "" {
		kid, err := s.searchIdentity(context.TODO(), req.Identity)
		if err != nil {
			return nil, err
		}
		ok, _, err := s.pull(ctx, kid)
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
	spks, err := s.vault.EdX25519PublicKeys()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load keys")
	}
	for _, spk := range spks {
		ok, _, err := s.pull(ctx, spk.ID())
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

func (s *service) pull(ctx context.Context, kid keys.ID) (bool, *user.Result, error) {
	logger.Infof("Pull %s", kid)

	if err := s.importID(kid); err != nil {
		return false, nil, err
	}

	ok, res, err := s.update(ctx, kid)
	if err != nil {
		return false, nil, err
	}
	return ok, res, nil
}

func (s *service) update(ctx context.Context, kid keys.ID) (bool, *user.Result, error) {
	logger.Infof("Update %s", kid)

	resp, err := s.remote.Sigchain(ctx, kid)
	if err != nil {
		return false, nil, err
	}
	if resp == nil {
		logger.Infof("No sigchain for %s", kid)
		return false, nil, nil
	}
	// TODO: Check that our existing statements haven't changed or disappeared
	logger.Infof("Received sigchain %s, len=%d", kid, len(resp.Statements))
	for _, st := range resp.Statements {
		b, err := st.Bytes()
		if err != nil {
			return false, nil, err
		}
		logger.Debugf("Saving %s %d", st.KID, st.Seq)
		if err := s.db.Set(ctx, docs.Path("sigchain", st.Key()), b); err != nil {
			return false, nil, err
		}
	}

	res, err := s.users.Update(ctx, kid)
	if err != nil {
		return false, nil, err
	}

	return true, res, nil
}
