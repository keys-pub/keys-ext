package service

import (
	"bytes"
	"context"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// Push (RPC) publishes user public key and sigchain.
func (s *service) Push(ctx context.Context, req *PushRequest) (*PushResponse, error) {
	key, err := s.parseKeyOrCurrent(req.KID)
	if err != nil {
		return nil, err
	}

	urls, err := s.push(key.ID())
	if err != nil {
		return nil, err
	}
	ok, err := s.pull(ctx, key.ID())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.Errorf("sigchain not found on remote after push")
	}

	return &PushResponse{
		KID:  key.ID().String(),
		URLs: urls,
	}, nil
}

func (s *service) publish(kid keys.ID) error {
	logger.Infof("Publishing %s", kid)

	// Local sigchain
	sc, err := s.scs.Sigchain(kid)
	if err != nil {
		return err
	}
	if sc == nil {
		return keys.NewErrNotFound(kid, keys.SigchainType)
	}
	sts := sc.Statements()
	if len(sts) == 0 {
		return errors.Errorf("no sigchain statements")
	}

	// Remote sigchain
	if s.remote == nil {
		return errors.Errorf("no remote set")
	}
	rsc, rerr := s.remote.Sigchain(kid)
	if rerr != nil {
		return rerr
	}
	if rsc != nil && len(rsc.Statements) > 0 {
		return errors.Errorf("sigchain already published")
	}
	for _, st := range sts {
		if err := s.remote.PutSigchainStatement(st); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) push(kid keys.ID) ([]string, error) {
	logger.Infof("Pushing %s", kid)

	// Local sigchain
	sc, err := s.scs.Sigchain(kid)
	if err != nil {
		return nil, err
	}
	if sc == nil {
		return nil, keys.NewErrNotFound(kid, keys.SigchainType)
	}
	sts := sc.Statements()

	// Remote sigchain
	if s.remote == nil {
		return nil, errors.Errorf("no remote set")
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
		url := s.remote.URL().String() + st.URLPath()
		urls = append(urls, url)
	}

	return urls, nil
}
