package service

import (
	"context"
	"strings"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// LookupOpts are options for key lookups.
type LookupOpts struct {
	// Verify to make sure the key and user is verified.
	// If not verified an error is returned.
	Verify bool
	// SearchRemote to look for the key on the server if not found locally.
	SearchRemote bool
}

// lookup a key by kid or user@service.
func (s *service) lookup(ctx context.Context, key string, opts *LookupOpts) (keys.ID, error) {
	if key == "" {
		return "", errors.Errorf("no key specified")
	}
	if opts == nil {
		opts = &LookupOpts{}
	}

	kid, err := s.lookupKID(ctx, key, opts.SearchRemote)
	if err != nil {
		return "", err
	}

	if opts.Verify {
		if err := s.ensureUsersVerified(ctx, kid); err != nil {
			return "", err
		}
	}

	return kid, nil
}

func (s *service) lookupAll(ctx context.Context, ks []string, opts *LookupOpts) ([]keys.ID, error) {
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

	return kid, nil
}
