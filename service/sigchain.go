package service

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
)

// Sigchain (RPC) ...
func (s *service) Sigchain(ctx context.Context, req *SigchainRequest) (*SigchainResponse, error) {
	kid, err := s.parseKID(req.KID)
	if err != nil {
		return nil, err
	}

	key, err := s.key(ctx, kid)
	if err != nil {
		return nil, err
	}

	sc, err := s.scs.Sigchain(kid)
	if err != nil {
		return nil, err
	}

	stsOut := statementsToRPC(sc.Statements())

	return &SigchainResponse{
		Key:        key,
		Statements: stsOut,
	}, nil
}

func statementFromRPC(st *Statement) (*keys.Statement, error) {
	ts := tsutil.ConvertMillis(st.Timestamp)
	kid, err := keys.ParseID(st.KID)
	if err != nil {
		return nil, err
	}
	return keys.NewUnverifiedStatement(st.Sig, st.Data, kid, int(st.Seq), st.Prev, int(st.Revoke), st.Type, ts)
}

// statementsFromRPC converts Statement's to keys.Statement's.
func statementsFromRPC(sts []*Statement) ([]*keys.Statement, error) {
	stsOut := make([]*keys.Statement, 0, len(sts))
	for _, st := range sts {
		stOut, err := statementFromRPC(st)
		if err != nil {
			return nil, err
		}
		stsOut = append(stsOut, stOut)
	}
	return stsOut, nil
}

func statementToRPC(st *keys.Statement) *Statement {
	return &Statement{
		Sig:       st.Sig,
		Data:      st.Data,
		KID:       st.KID.String(),
		Seq:       int32(st.Seq),
		Prev:      st.Prev,
		Revoke:    int32(st.Revoke),
		Timestamp: tsutil.Millis(st.Timestamp),
		Type:      st.Type,
	}
}

func statementsToRPC(sts []*keys.Statement) []*Statement {
	stsOut := make([]*Statement, 0, len(sts))
	for _, st := range sts {
		stsOut = append(stsOut, statementToRPC(st))
	}
	return stsOut
}

func sigchainFromRPC(kid keys.ID, ssts []*Statement) (*keys.Sigchain, error) {
	logger.Infof("Resolving sigchain from statements")
	sc := keys.NewSigchain(kid)
	sts, err := statementsFromRPC(ssts)
	if err != nil {
		return nil, err
	}
	if err := sc.AddAll(sts); err != nil {
		return nil, errors.Wrapf(err, "failed to resolve sigchain from statements")
	}
	return sc, nil
}

// Sigchain (RPC) ...
func (s *service) Statement(ctx context.Context, req *StatementRequest) (*StatementResponse, error) {
	kid, err := s.parseKID(req.KID)
	if err != nil {
		return nil, err
	}

	sc, err := s.scs.Sigchain(kid)
	if err != nil {
		return nil, err
	}
	if sc == nil {
		return nil, keys.NewErrNotFound(kid.String())
	}
	stsOut := statementsToRPC(sc.Statements())

	if req.Seq < 0 {
		return nil, errors.Errorf("invalid seq")
	}
	if req.Seq == 0 {
		return nil, errors.Errorf("no seq specified")
	}
	if int(req.Seq-1) >= len(stsOut) {
		return nil, errors.Errorf("seq too large")
	}
	return &StatementResponse{
		Statement: stsOut[req.Seq-1],
	}, nil
}

// StatementCreate (RPC) ...
func (s *service) StatementCreate(ctx context.Context, req *StatementCreateRequest) (*StatementCreateResponse, error) {
	if req.KID == "" {
		return nil, errors.Errorf("no kid specified")
	}
	kid, err := keys.ParseID(req.KID)
	if err != nil {
		return nil, err
	}
	key, err := s.edX25519Key(kid)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, keys.NewErrNotFound(req.KID)
	}

	sc, err := s.scs.Sigchain(key.ID())
	if err != nil {
		return nil, err
	}
	st, err := keys.NewSigchainStatement(sc, req.Data, key, "", s.clock.Now())
	if err != nil {
		return nil, err
	}
	if err := sc.Add(st); err != nil {
		return nil, err
	}

	if !req.Local {
		if err := s.client.SigchainSave(ctx, st); err != nil {
			return nil, err
		}
	}

	if err := s.scs.Save(sc); err != nil {
		return nil, err
	}

	stOut := statementToRPC(st)

	return &StatementCreateResponse{
		Statement: stOut,
	}, nil
}

// StatementRevoke (RPC) ...
func (s *service) StatementRevoke(ctx context.Context, req *StatementRevokeRequest) (*StatementRevokeResponse, error) {
	kid, err := keys.ParseID(req.KID)
	if err != nil {
		return nil, err
	}
	key, err := s.edX25519Key(kid)
	if err != nil {
		return nil, err
	}

	sc, err := s.scs.Sigchain(key.ID())
	if err != nil {
		return nil, err
	}

	st, err := sc.Revoke(int(req.Seq), key)
	if err != nil {
		return nil, err
	}

	if !req.Local {
		if err := s.client.SigchainSave(ctx, st); err != nil {
			return nil, err
		}
	}

	if err := s.scs.Save(sc); err != nil {
		return nil, err
	}

	if _, err = s.users.Update(ctx, key.ID()); err != nil {
		return nil, err
	}

	stOut := statementToRPC(st)

	return &StatementRevokeResponse{
		Statement: stOut,
	}, nil
}
