package server

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/api"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

const sigchainChanges = "sigchain-changes"

func (s *Server) listSigchains(c echo.Context) error {
	ctx := c.Request().Context()
	logger.Infof(ctx, "Server GET sigchains %s", s.urlWithBase(c))

	var version time.Time
	if f := c.QueryParam("version"); f != "" {
		i, err := strconv.Atoi(f)
		if err != nil {
			return err
		}
		version = keys.TimeFromMillis(keys.TimeMs(i))
	}
	plimit := c.QueryParam("limit")
	if plimit == "" {
		plimit = "100"
	}
	limit, err := strconv.Atoi(plimit)
	if err != nil {
		return ErrBadRequest(c, errors.Wrapf(err, "invalid limit"))
	}
	if limit < 1 {
		return ErrBadRequest(c, errors.Errorf("invalid limit, too small"))
	}
	if limit > 100 {
		return ErrBadRequest(c, errors.Errorf("invalid limit, too large"))
	}

	changes, to, err := s.fi.Changes(ctx, sigchainChanges, version, limit, keys.Ascending)
	if err != nil {
		return internalError(c, err)
	}
	paths := make([]string, 0, len(changes))
	for _, a := range changes {
		paths = append(paths, a.Path)
	}
	md := make(map[string]api.Metadata, len(paths))
	statements := make([]*keys.Statement, 0, len(paths))

	docs, err := s.fi.GetAll(ctx, paths)
	if err != nil {
		return internalError(c, err)
	}

	for _, doc := range docs {
		st, err := s.statementFromBytes(ctx, doc.Data)
		if err != nil {
			return internalError(c, err)
		}

		statements = append(statements, st)

		md[st.URL()] = api.Metadata{
			CreatedAt: doc.CreatedAt,
			UpdatedAt: doc.UpdatedAt,
		}
	}
	versionNext := fmt.Sprintf("%d", keys.TimeToMillis(to))
	resp := api.SigchainsResponse{
		Statements: statements,
		Version:    versionNext,
	}
	fields := keys.NewStringSetSplit(c.QueryParam("include"), ",")
	if fields.Contains("md") {
		resp.Metadata = md
	}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) sigchain(c echo.Context, kid keys.ID) (*keys.Sigchain, map[string]api.Metadata, error) {
	ctx := c.Request().Context()
	iter, err := s.fi.Documents(ctx, SigchainResource.String(), &keys.DocumentsOpts{Prefix: kid.String()})
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
		st, err := keys.StatementFromBytes(doc.Data)
		if err != nil {
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
	ctx := c.Request().Context()
	logger.Infof(ctx, "Server GET sigchain %s", s.urlWithBase(c))

	kid, err := keys.ParseID(c.Param("kid"))
	if err != nil {
		return ErrNotFound(c, nil)
	}

	logger.Infof(ctx, "Loading sigchain: %s", kid)
	sc, md, err := s.sigchain(c, kid)
	if err != nil {
		return internalError(c, err)
	}
	if sc.Length() == 0 {
		return ErrNotFound(c, errors.Errorf("sigchain not found"))
	}
	resp := api.SigchainResponse{
		KID:        kid,
		Statements: sc.Statements(),
	}
	fields := keys.NewStringSetSplit(c.QueryParam("include"), ",")
	if fields.Contains("md") {
		resp.Metadata = md
	}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) getSigchainStatement(c echo.Context) error {
	ctx := c.Request().Context()
	logger.Infof(ctx, "Server GET sigchain statement %s", c.Path())

	kid, err := keys.ParseID(c.Param("kid"))
	if err != nil {
		return ErrBadRequest(c, err)
	}
	i, err := strconv.Atoi(c.Param("seq"))
	if err != nil {
		return internalError(c, err)
	}
	path := keys.Path(SigchainResource, kid.WithSeq(i))
	st, doc, err := s.statement(ctx, path)
	if st == nil {
		return ErrNotFound(c, errors.Errorf("statement not found"))
	}
	if err != nil {
		return internalError(c, err)
	}
	if !doc.CreatedAt.IsZero() {
		c.Response().Header().Set("CreatedAt", doc.CreatedAt.Format(http.TimeFormat))
		c.Response().Header().Set("CreatedAt-RFC3339M", doc.CreatedAt.Format(keys.RFC3339Milli))
	}
	if !doc.UpdatedAt.IsZero() {
		c.Response().Header().Set("Last-Modified", doc.UpdatedAt.Format(http.TimeFormat))
		c.Response().Header().Set("Last-Modified-RFC3339M", doc.UpdatedAt.Format(keys.RFC3339Milli))
	}

	return JSON(c, http.StatusOK, st)
}

func (s *Server) putSigchainStatement(c echo.Context) error {
	ctx := c.Request().Context()
	logger.Infof(ctx, "Server PUT sigchain statement %s", s.urlWithBase(c))

	if c.Request().Body == nil {
		return ErrBadRequest(c, errors.Errorf("missing body"))
	}

	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return internalError(c, err)
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

	path := keys.Path(SigchainResource, st.Key())

	exists, err := s.fi.Exists(ctx, path)
	if err != nil {
		return internalError(c, err)
	}
	if exists {
		return ErrConflict(c, errors.Errorf("statement already exists"))
	}

	if access := s.accessFn(c, SigchainResource, Put); !access.Allow {
		return ErrResponse(c, access.StatusCode, access.Message)
	}

	sc, _, err := s.sigchain(c, st.KID)
	if err != nil {
		return internalError(c, err)
	}

	if sc.Length() >= 128 {
		// TODO: Increase limits
		return ErrEntityTooLarge(c, errors.Errorf("sigchain limit reached, contact gabriel@github to bump the limits"))
	}

	prev := sc.Last()
	if err := sc.VerifyStatement(st, prev); err != nil {
		return ErrBadRequest(c, err)
	}

	logger.Infof(ctx, "Statement, set %s", path)
	if err := s.fi.Create(ctx, path, b); err != nil {
		return internalError(c, err)
	}
	if err := s.fi.ChangeAdd(ctx, sigchainChanges, path); err != nil {
		return internalError(c, err)
	}

	if err := s.tasks.CreateTask(ctx, "POST", "/task/check/"+st.KID.String(), s.internalAuth); err != nil {
		return internalError(c, err)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) statement(ctx context.Context, path string) (*keys.Statement, *keys.Document, error) {
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
	st, err := keys.StatementFromBytes(b)
	if err != nil {
		return nil, err
	}
	bout := st.Bytes()
	if !bytes.Equal(b, bout) {
		logger.Debugf(context.TODO(), "%s != %s", string(b), string(bout))
		return nil, errors.Errorf("invalid statement bytes")
	}
	if err := st.Verify(); err != nil {
		return st, err
	}
	return st, nil
}
