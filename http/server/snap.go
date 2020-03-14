package server

import (
	"io/ioutil"
	"net/http"

	"github.com/keys-pub/keys"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) putSnap(c echo.Context) error {
	request := c.Request()
	ctx := request.Context()
	logger.Infof(ctx, "Server PUT snap %s", s.urlString(c))

	kid, status, err := s.authorize(c)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	if c.Request().Body == nil {
		return ErrBadRequest(c, errors.Errorf("missing body"))
	}

	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return internalError(c, err)
	}

	path := keys.Path("snap", kid)
	logger.Infof(ctx, "Save snap %s", path)
	if err := s.fi.Set(ctx, path, b); err != nil {
		return internalError(c, err)
	}

	return c.String(http.StatusOK, "{}")
}

func (s *Server) getSnap(c echo.Context) error {
	request := c.Request()
	ctx := request.Context()
	logger.Infof(ctx, "Server GET snap %s", s.urlString(c))

	kid, status, err := s.authorize(c)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	path := keys.Path("snap", kid)
	logger.Infof(ctx, "Get snap %s", path)
	doc, err := s.fi.Get(ctx, path)
	if err != nil {
		return internalError(c, err)
	}
	if doc == nil {
		return ErrNotFound(c, errors.Errorf("snap not found"))
	}

	return c.Blob(http.StatusOK, "", doc.Data)
}

func (s *Server) deleteSnap(c echo.Context) error {
	request := c.Request()
	ctx := request.Context()
	logger.Infof(ctx, "Server DELETE snap %s", s.urlString(c))

	kid, status, err := s.authorize(c)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	path := keys.Path("snap", kid)
	logger.Infof(ctx, "Delete snap %s", path)
	if _, err := s.fi.Delete(ctx, path); err != nil {
		return internalError(c, err)
	}
	return c.String(http.StatusOK, "{}")
}
