package service

import (
	"bytes"
	"context"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// Push (RPC) publishes sigchain statements.
func (s *service) Push(ctx context.Context, req *PushRequest) (*PushResponse, error) {
	key, err := s.parseSignKey(req.KID, true)
	if err != nil {
		return nil, err
	}

	urls, err := s.push(ctx, key.ID())
	if err != nil {
		return nil, err
	}

	if len(urls) == 0 {
		return nil, errors.Errorf("nothing to push")
	}

	return &PushResponse{
		KID:  key.ID().String(),
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

	rsc, rerr := s.remote.Sigchain(kid)
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
			if !bytes.Equal(st.Bytes(), rsts[i].Bytes()) {
				return nil, errors.Errorf("remote and local sigchain statements differ")
			}
		} else {
			if err := s.remote.PutSigchainStatement(st); err != nil {
				return nil, err
			}
		}
		url := s.remote.URL().String() + st.URL()
		urls = append(urls, url)
	}

	// TODO: instead of pulling, save resource from push
	if _, err := s.pull(ctx, kid); err != nil {
		return nil, err
	}

	return urls, nil
}
