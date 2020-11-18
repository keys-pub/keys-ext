package service

import (
	"context"
	"strings"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// lookupOpts are options for key lookups.
type lookupOpts struct {
	// Verify makes sure the key and user is verified and the verification is up to date.
	Verify bool
	// SearchRemote to look for the key on the server if not found locally.
	SearchRemote bool
}

// lookup a key by kid or user@service.
func (s *service) lookup(ctx context.Context, key string, opts *lookupOpts) (keys.ID, error) {
	if key == "" {
		return "", errors.Errorf("no key specified")
	}
	if opts == nil {
		opts = &lookupOpts{}
	}

	kid, err := s.lookupKID(ctx, key, opts.SearchRemote)
	if err != nil {
		return "", err
	}
	if kid == "" {
		return "", keys.NewErrNotFound(key)
	}

	if opts.Verify {
		if err := s.ensureUserVerified(ctx, kid); err != nil {
			return "", err
		}
	}

	return kid, nil
}

func (s *service) lookupAll(ctx context.Context, ks []string, opts *lookupOpts) ([]keys.ID, error) {
	ids := make([]keys.ID, 0, len(ks))
	for _, key := range ks {
		id, err := s.lookup(ctx, key, opts)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (s *service) lookupKID(ctx context.Context, key string, searchRemote bool) (keys.ID, error) {
	if strings.Contains(key, "@") {
		return s.lookupUser(ctx, key, searchRemote)
	}

	kid, err := keys.ParseID(key)
	if err != nil {
		return "", errors.Errorf("failed to parse %s", key)
	}

	rkid, err := s.scs.Lookup(kid)
	if err != nil {
		return "", err
	}
	if rkid != "" {
		kid = rkid
	}

	if searchRemote {
		res, err := s.client.User(ctx, kid)
		if err != nil {
			return "", err
		}
		if res != nil {
			// TODO: We should verify ourselves that the server isn't lying
			return res.User.KID, nil
		}
	}
	return kid, nil
}
