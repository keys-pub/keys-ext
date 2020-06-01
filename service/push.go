package service

import (
	"bytes"
	"context"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// Push (RPC) publishes sigchain statements.
func (s *service) Push(ctx context.Context, req *PushRequest) (*PushResponse, error) {
	kid, err := s.parseIdentity(context.TODO(), req.Identity, false)
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
		ks := s.keyStore()
		key, err := ks.EdX25519Key(kid)
		if err != nil {
			return nil, err
		}
		if key == nil {
			return nil, keys.NewErrNotFound(kid.String())
		}
		if err := s.remote.Check(ctx, key); err != nil {
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

	rsc, rerr := s.remote.Sigchain(ctx, kid)
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
			if err := s.remote.PutSigchainStatement(ctx, st); err != nil {
				return nil, err
			}
		}
		url := s.remote.URL().String() + st.URL()
		urls = append(urls, url)
	}

	// TODO: instead of pulling, save resource from push
	if _, _, err := s.pull(ctx, kid); err != nil {
		return nil, err
	}

	return urls, nil
}
