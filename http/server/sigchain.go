package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/tsutil"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) sigchain(c echo.Context, kid keys.ID) (*keys.Sigchain, map[string]api.Metadata, error) {
	ctx := c.Request().Context()
	iter, err := s.fi.DocumentIterator(ctx, SigchainResource.String(), ds.Prefix(kid.String()))
	defer iter.Release()
	if err != nil {
		return nil, nil, err
	}
	sc := keys.NewSigchain(kid)
	md := make(map[string]api.Metadata, 100)
	for {
		doc, err := iter.Next()
		if err != nil {
			return nil, nil, err
		}
		if doc == nil {
			break
		}

		var st *keys.Statement
		if err := json.Unmarshal(doc.Data, &st); err != nil {
			return nil, nil, err
		}

		if err := sc.Add(st); err != nil {
			return nil, nil, err
		}
		md[st.URL()] = api.Metadata{
			CreatedAt: doc.CreatedAt,
			UpdatedAt: doc.UpdatedAt,
		}
	}

	return sc, md, nil
}

func (s *Server) getSigchain(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	kid, err := keys.ParseID(c.Param("kid"))
	if err != nil {
		return ErrNotFound(c, nil)
	}

	s.logger.Infof("Loading sigchain: %s", kid)
	sc, md, err := s.sigchain(c, kid)
	if err != nil {
		return s.internalError(c, err)
	}
	if sc.Length() == 0 {
		return ErrNotFound(c, errors.Errorf("sigchain not found"))
	}
	resp := api.SigchainResponse{
		KID:        kid,
		Statements: sc.Statements(),
	}
	fields := ds.NewStringSetSplit(c.QueryParam("include"), ",")
	if fields.Contains("md") {
		resp.Metadata = md
	}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) getSigchainStatement(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	kid, err := keys.ParseID(c.Param("kid"))
	if err != nil {
		return ErrNotFound(c, err)
	}
	i, err := strconv.Atoi(c.Param("seq"))
	if err != nil {
		return ErrNotFound(c, err)
	}
	path := ds.Path(SigchainResource, kid.WithSeq(i))
	st, doc, err := s.statement(ctx, path)
	if st == nil {
		return ErrNotFound(c, errors.Errorf("statement not found"))
	}
	if err != nil {
		return s.internalError(c, err)
	}
	if !doc.CreatedAt.IsZero() {
		c.Response().Header().Set("CreatedAt", doc.CreatedAt.Format(http.TimeFormat))
		c.Response().Header().Set("CreatedAt-RFC3339M", doc.CreatedAt.Format(tsutil.RFC3339Milli))
	}
	if !doc.UpdatedAt.IsZero() {
		c.Response().Header().Set("Last-Modified", doc.UpdatedAt.Format(http.TimeFormat))
		c.Response().Header().Set("Last-Modified-RFC3339M", doc.UpdatedAt.Format(tsutil.RFC3339Milli))
	}

	return JSON(c, http.StatusOK, st)
}

func (s *Server) putSigchainStatement(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	if c.Request().Body == nil {
		return ErrBadRequest(c, errors.Errorf("missing body"))
	}

	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return s.internalError(c, err)
	}
	st, err := s.statementFromBytes(ctx, b)
	if err != nil {
		return ErrBadRequest(c, err)
	}
	if len(st.Data) > 16*1024 {
		return ErrBadRequest(c, errors.Errorf("too much data for sigchain statement (greater than 16KiB)"))
	}

	if c.Param("kid") != st.KID.String() {
		return ErrBadRequest(c, errors.Errorf("invalid kid"))
	}
	if c.Param("seq") != fmt.Sprintf("%d", st.Seq) {
		return ErrBadRequest(c, errors.Errorf("invalid seq"))
	}

	path := ds.Path(SigchainResource, st.Key())

	exists, err := s.fi.Exists(ctx, path)
	if err != nil {
		return s.internalError(c, err)
	}
	if exists {
		return ErrConflict(c, errors.Errorf("statement already exists"))
	}

	if access := s.accessFn(c, SigchainResource, Put); !access.Allow {
		return ErrResponse(c, access.StatusCode, access.Message)
	}

	sc, _, err := s.sigchain(c, st.KID)
	if err != nil {
		return s.internalError(c, err)
	}

	if sc.Length() >= 128 {
		// TODO: Increase limits
		return ErrEntityTooLarge(c, errors.Errorf("sigchain limit reached, contact gabriel@github to bump the limits"))
	}

	prev := sc.Last()
	if err := sc.VerifyStatement(st, prev); err != nil {
		return ErrBadRequest(c, err)
	}
	if err := sc.Add(st); err != nil {
		return ErrBadRequest(c, err)
	}

	// Check we don't have an existing user with a different key, which would cause duplicates in search.
	// They should revoke the existing user before linking a new key.
	// Since there is a delay in indexing this won't stop a malicious user from creating duplicates but
	// it will limit them. If we find spaming this is a problem, we can get more strict.
	existing, err := s.users.CheckForExisting(ctx, sc)
	if err != nil {
		return s.internalError(c, err)
	}
	if existing != "" {
		if err := s.checkKID(ctx, st.KID); err != nil {
			return s.internalError(c, err)
		}
		return ErrResponse(c, http.StatusConflict, fmt.Sprintf("user already exists with key %s, if you removed or revoked the previous statement you may need to wait briefly for search to update", existing))
	}

	s.logger.Infof("Statement, set %s", path)
	if err := s.fi.Create(ctx, path, b); err != nil {
		return s.internalError(c, err)
	}

	if err := s.tasks.CreateTask(ctx, "POST", "/task/check/"+st.KID.String(), s.internalAuth); err != nil {
		return s.internalError(c, err)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) statement(ctx context.Context, path string) (*keys.Statement, *ds.Document, error) {
	e, err := s.fi.Get(ctx, path)
	if err != nil {
		return nil, nil, err
	}
	if e == nil {
		return nil, nil, nil
	}
	st, err := s.statementFromBytes(ctx, e.Data)
	if err != nil {
		return nil, nil, err
	}
	return st, e, nil
}

func (s *Server) statementFromBytes(ctx context.Context, b []byte) (*keys.Statement, error) {
	var st *keys.Statement
	if err := json.Unmarshal(b, &st); err != nil {
		return nil, err
	}
	bout, err := st.Bytes()
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(b, bout) {
		s.logger.Errorf("%s != %s", string(b), string(bout))
		return nil, errors.Errorf("invalid statement bytes")
	}
	if err := st.Verify(); err != nil {
		return st, err
	}
	return st, nil
}

func (s *Server) getSigchainAliased(c echo.Context) error {
	if c.Request().Host == "sigcha.in" {
		return s.getSigchain(c)
	}
	return ErrNotFound(c, nil)

}
func (s *Server) getSigchainStatementAliased(c echo.Context) error {
	if c.Request().Host == "sigcha.in" {
		return s.getSigchainStatement(c)
	}
	return ErrNotFound(c, nil)
}

func (s *Server) putSigchainStatementAliased(c echo.Context) error {
	// if c.Request().Host == "sigcha.in" {
	// 	return s.putSigchainStatement(c)
	// }
	// return ErrNotFound(c, nil)

	// http/client doesn't specify /sigchain/:kid/:seq path on earlier versions of the app.
	return s.putSigchainStatement(c)
}
