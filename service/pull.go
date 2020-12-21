package service

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/user"
	"github.com/pkg/errors"
)

// Pull (RPC) imports a key from the server and updates it.
// If no key is specified, we update our existing keys.
func (s *service) Pull(ctx context.Context, req *PullRequest) (*PullResponse, error) {
	if req.Key != "" {
		kid, err := s.lookup(ctx, req.Key, &lookupOpts{SearchRemote: true})
		if err != nil {
			return nil, err
		}
		res, err := s.pullUser(ctx, kid)
		if err != nil {
			return nil, err
		}
		if res == nil {
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
		res, err := s.pullUser(ctx, spk.ID())
		if err != nil {
			return nil, err
		}
		if res == nil {
			// TODO: Report missing
			continue
		}
		pulled = append(pulled, spk.ID().String())
	}
	return &PullResponse{KIDs: pulled}, nil
}

func (s *service) pullUser(ctx context.Context, kid keys.ID) (*user.Result, error) {
	logger.Infof("Pull user %s", kid)
	if err := s.importID(kid); err != nil {
		return nil, err
	}
	res, err := s.updateUser(ctx, kid, false)
	if err != nil {
		return nil, err
	}
	return res, nil
}
