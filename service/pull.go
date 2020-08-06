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
		kid, err := s.lookup(ctx, req.Key, &LookupOpts{SearchRemote: true})
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
