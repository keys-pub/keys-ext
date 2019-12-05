package server

import (
	"io"
	"io/ioutil"
	"net/http"

	"github.com/keys-pub/keys"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) putShare(c echo.Context) error {
	body := c.Request().Body
	return s.putShareBody(c, body)
}

func (s *Server) putShareBody(c echo.Context, body io.Reader) error {
	request := c.Request()
	ctx := request.Context()
	logger.Infof(ctx, "Server PUT share %s", s.urlString(c))

	// Auth
	auth := request.Header.Get("Authorization")
	if auth == "" {
		return ErrUnauthorized(c, errors.Errorf("missing Authorization header"))
	}
	now := s.nowFn()
	authRes, err := CheckAuthorization(request.Context(), request.Method, s.urlString(c), auth, s.mc, now)
	if err != nil {
		return ErrForbidden(c, err)
	}

	kid, err := keys.ParseID(c.Param("kid"))
	if err != nil {
		return ErrBadRequest(c, err)
	}

	if authRes.kid != kid {
		return ErrForbidden(c, errors.Errorf("invalid kid"))
	}

	recipient, err := keys.ParseID(c.Param("recipient"))
	if err != nil {
		return ErrBadRequest(c, err)
	}

	// TODO: Support destroy after time limit or access

	path := keys.Path("share-"+kid, recipient)

	// If not body, remove only and return
	if body == nil {
		ok, err := s.fi.Delete(ctx, path)
		if err != nil {
			return internalError(c, err)
		}
		if !ok {
			return ErrNotFound(c, errors.Errorf("share not found"))
		}
		return c.String(http.StatusOK, "{}")
	}

	// If exists, remove and overwrite
	exists, err := s.fi.Exists(ctx, path)
	if err != nil {
		return internalError(c, err)
	}
	if exists {
		ok, err := s.fi.Delete(ctx, path)
		if err != nil {
			return internalError(c, err)
		}
		if !ok {
			return ErrNotFound(c, errors.Errorf("share not found"))
		}
	}

	bin, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return internalError(c, err)
	}

	if err := s.fi.Create(ctx, path, bin); err != nil {
		return internalError(c, err)
	}

	return c.String(http.StatusOK, "{}")
}

func (s *Server) getShare(c echo.Context) error {
	request := c.Request()
	ctx := request.Context()
	logger.Infof(ctx, "Server GET share %s", s.urlString(c))

	// Auth
	auth := request.Header.Get("Authorization")
	if auth == "" {
		return ErrUnauthorized(c, errors.Errorf("missing Authorization header"))
	}
	now := s.nowFn()
	authRes, err := CheckAuthorization(request.Context(), request.Method, s.urlString(c), auth, s.mc, now)
	if err != nil {
		return ErrForbidden(c, err)
	}

	recipient, err := keys.ParseID(c.Param("recipient"))
	if err != nil {
		return ErrBadRequest(c, err)
	}

	if authRes.kid != recipient {
		return ErrForbidden(c, errors.Errorf("invalid kid"))
	}

	kid, err := keys.ParseID(c.Param("kid"))
	if err != nil {
		return ErrBadRequest(c, err)
	}

	// TODO: Support destroy after time limit or access?

	e, err := s.fi.Get(ctx, keys.Path("share-"+kid, recipient))
	if err != nil {
		return internalError(c, err)
	}
	if e == nil {
		return ErrNotFound(c, errors.Errorf("share not found"))
	}

	return c.String(http.StatusOK, string(e.Data))
}

func (s *Server) deleteShare(c echo.Context) error {
	logger.Infof(c.Request().Context(), "Server DELETE share %s", s.urlString(c))
	return s.putShareBody(c, nil)
}
