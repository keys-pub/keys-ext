package server

import (
	"net/http"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) postDrop(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	body, st, err := readBody(c, true, 1024)
	if err != nil {
		return ErrResponse(c, st, err)
	}

	kid, err := keys.ParseID(c.Param("kid"))
	if err != nil {
		return ErrBadRequest(c, errors.Errorf("invalid drop kid"))
	}

	token := c.QueryParam("token")
	if token != "" {
		hasToken, err := s.checkUserToken(ctx, token, kid)
		if err != nil {
			return s.internalError(c, err)
		}
		if !hasToken {
			return ErrForbidden(c, errors.Errorf("invalid token"))
		}
		path := dstore.Path("pdrops", kid)
		if _, _, err := s.fi.EventsAdd(ctx, path, [][]byte{body}); err != nil {
			return err
		}
		var out struct{}
		return JSON(c, http.StatusOK, out)
	}

	path := dstore.Path("drops", kid)
	if _, _, err := s.fi.EventsAdd(ctx, path, [][]byte{body}); err != nil {
		return err
	}
	var out struct{}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) getDrops(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	auth, err := s.auth(c, newAuth("Authorization", "kid", nil))
	if err != nil {
		return ErrForbidden(c, err)
	}

	path := dstore.Path("drops", auth.KID)
	resp, err := s.events(c, path, 0)
	if err != nil {
		return s.internalError(c, err)
	}
	out := &api.DropsResponse{Drops: resp.Events}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) deleteDrops(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	auth, err := s.auth(c, newAuth("Authorization", "kid", nil))
	if err != nil {
		return ErrForbidden(c, err)
	}

	path := dstore.Path("drops", auth.KID)
	ok, err := s.fi.EventsDelete(ctx, path)
	if err != nil {
		return s.internalError(c, err)
	}
	if !ok {
		return ErrNotFound(c, errors.Errorf("drop not found"))
	}

	var out struct{}
	return JSON(c, http.StatusOK, out)
}
