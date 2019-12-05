package service

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// Sigchain (RPC) ...
func (s *service) Sigchain(ctx context.Context, req *SigchainRequest) (*SigchainResponse, error) {
	kid, err := s.parseKIDOrCurrent(req.KID)
	if err != nil {
		return nil, err
	}

	sc, err := s.scs.Sigchain(kid)
	if err != nil {
		return nil, err
	}
	if sc == nil {
		return nil, keys.NewErrNotFound(kid, keys.SigchainType)
	}
	stsOut := statementsToRPC(sc.Statements())

	if req.Seq != 0 {
		if int(req.Seq-1) >= len(stsOut) {
			return nil, errors.Errorf("seq too large")
		}
		return &SigchainResponse{
			Statements: []*Statement{stsOut[req.Seq-1]},
		}, nil
	}

	return &SigchainResponse{
		KID:        kid.String(),
		Statements: stsOut,
	}, nil
}

func statementFromRPC(st *Statement) (*keys.Statement, error) {
	kid, err := keys.ParseID(st.KID)
	if err != nil {
		return nil, err
	}
	ts := keys.TimeFromMillis(keys.TimeMs(st.Timestamp))
	return keys.NewStatement(st.Sig, st.Data, kid, int(st.Seq), st.Prev, int(st.Revoke), st.Type, ts)
}

// statementsFromRPC converts Statement's to keys.Statement's.
func statementsFromRPC(sts []*Statement) ([]*keys.Statement, error) {
	stsOut := make([]*keys.Statement, 0, len(sts))
	for _, st := range sts {
		stOut, stOutErr := statementFromRPC(st)
		if stOutErr != nil {
			return nil, stOutErr
		}
		stsOut = append(stsOut, stOut)
	}
	return stsOut, nil
}

func statementToRPC(st *keys.Statement) *Statement {
	return &Statement{
		Sig:    st.Sig,
		Data:   st.Data,
		KID:    st.KID.String(),
		Seq:    int32(st.Seq),
		Prev:   st.Prev,
		Revoke: int32(st.Revoke),
		Type:   st.Type,
	}
}

func statementsToRPC(sts []*keys.Statement) []*Statement {
	stsOut := make([]*Statement, 0, len(sts))
	for _, st := range sts {
		stsOut = append(stsOut, statementToRPC(st))
	}
	return stsOut
}

// SigchainStatementCreate (RPC) ...
func (s *service) SigchainStatementCreate(ctx context.Context, req *SigchainStatementCreateRequest) (*SigchainStatementCreateResponse, error) {
	key, err := s.parseKey(req.KID)
	if err != nil {
		return nil, err
	}

	sc, err := s.scs.Sigchain(key.ID())
	if err != nil {
		return nil, err
	}
	st, err := keys.GenerateStatement(sc, req.Data, key.SignKey(), "", s.Now())
	if err != nil {
		return nil, err
	}

	if !req.DryRun {
		if !req.Local {
			// TODO: Check sigchain status (local changes not pushed)

			// Save to remote
			if s.remote == nil {
				return nil, errors.Errorf("no remote set")
			}
			err := s.remote.PutSigchainStatement(st)
			if err != nil {
				return nil, err
			}
		}
		if err := s.scs.AddStatement(st, key.SignKey()); err != nil {
			return nil, err
		}
	}

	stOut := statementToRPC(st)

	return &SigchainStatementCreateResponse{
		Statement: stOut,
	}, nil
}

// SigchainStatementRevoke (RPC) ...
func (s *service) SigchainStatementRevoke(ctx context.Context, req *SigchainStatementRevokeRequest) (*SigchainStatementRevokeResponse, error) {
	return nil, errors.Errorf("not implemented")
}
