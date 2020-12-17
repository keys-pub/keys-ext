package service

import (
	"bytes"
	"context"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// Push (RPC) publishes sigchain statements.
func (s *service) Push(ctx context.Context, req *PushRequest) (*PushResponse, error) {
	kid, err := s.lookup(ctx, req.Key, &lookupOpts{VerifyUser: true})
	if err != nil {
		return nil, err
	}

	urls, err := s.push(ctx, kid)
	if err != nil {
		return nil, err
	}

	if len(urls) == 0 {
		return nil, errors.Errorf("nothing to push")
	}

	// TODO: Is remote check appropriate here?
	if req.RemoteCheck {
		key, err := s.vault.Key(kid)
		if err != nil {
			return nil, err
		}
		if key == nil {
			return nil, keys.NewErrNotFound(kid.String())
		}
		sk := key.AsEdX25519()
		if sk == nil {
			return nil, errors.Errorf("invalid key")
		}
		if err := s.client.Check(ctx, sk); err != nil {
			return nil, err
		}
	}

	return &PushResponse{
		KID:  kid.String(),
		URLs: urls,
	}, nil
}

func (s *service) push(ctx context.Context, kid keys.ID) ([]string, error) {
	logger.Infof("Pushing %s", kid)

	// Local sigchain
	sc, err := s.scs.Sigchain(kid)
	if err != nil {
		return nil, err
	}
	sts := sc.Statements()
	if len(sts) == 0 {
		return []string{}, nil
	}

	rsc, rerr := s.client.Sigchain(ctx, kid)
	if rerr != nil {
		return nil, rerr
	}

	var rsts []*keys.Statement
	if rsc != nil {
		rsts = rsc.Statements
	}

	urls := make([]string, 0, len(sts))
	for i, st := range sts {
		if i < len(rsts) {
			b, err := st.Bytes()
			if err != nil {
				return nil, err
			}
			b2, err := rsts[i].Bytes()
			if err != nil {
				return nil, err
			}
			if !bytes.Equal(b, b2) {
				return nil, errors.Errorf("remote and local sigchain statements differ")
			}
		} else {
			if err := s.client.SigchainSave(ctx, st); err != nil {
				return nil, err
			}
		}
		url := s.client.URL().String() + st.URL()
		urls = append(urls, url)
	}

	// TODO: instead of pulling, save resource from push
	if _, err := s.pullUser(ctx, kid); err != nil {
		return nil, err
	}

	return urls, nil
}
