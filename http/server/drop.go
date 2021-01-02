package server

import (
	"net/http"
	"net/url"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type drop struct {
	Token string `json:"token"`
}

func (s *Server) putDropAuth(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	body, st, err := readBody(c, true, 64*1024)
	if err != nil {
		return s.ErrResponse(c, st, err)
	}

	auth, err := s.auth(c, newAuth("Authorization", "kid", body))
	if err != nil {
		return s.ErrForbidden(c, err)
	}

	form, err := url.ParseQuery(string(body))
	if err != nil {
		return s.ErrBadRequest(c, err)
	}
	token := form.Get("token")
	if token == "" {
		return s.ErrBadRequest(c, errors.Errorf("invalid token"))
	}

	path := dstore.Path("dms", auth.KID)

	drop := &drop{Token: token}
	if err := s.fi.Set(ctx, path, dstore.From(drop), dstore.MergeAll()); err != nil {
		return s.ErrInternalServer(c, err)
	}
	var out struct{}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) postDrop(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	kid, err := keys.ParseID(c.Param("kid"))
	if err != nil {
		return s.ErrBadRequest(c, errors.Errorf("invalid recipient"))
	}

	token := c.Request().Header.Get("Authorization")
	if token == "" {
		return s.ErrForbidden(c, errors.Errorf("missing Authorization (token)"))
	}

	path := dstore.Path("dms", kid)

	var drop drop
	if _, err := s.fi.Load(ctx, path, &drop); err != nil {
		return s.ErrInternalServer(c, err)
	}
	if token != drop.Token {
		return s.ErrForbidden(c, errors.Errorf("invalid token"))
	}

	body, st, err := readBody(c, true, 1024)
	if err != nil {
		return s.ErrResponse(c, st, err)
	}

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
		return s.ErrForbidden(c, err)
	}

	limit := 1000
	path := dstore.Path("dms", auth.KID)
	resp, st, err := s.events(c, path, limit)
	if err != nil {
		return s.ErrResponse(c, st, err)
	}

	truncated := false
	if len(resp.Events) >= limit {
		// TODO: This is a lie if the number of results are exactly equal to limit
		truncated = true
	}

	out := &api.MessagesResponse{
		Messages:  resp.Events,
		Index:     resp.Index,
		Truncated: truncated,
	}

	return JSON(c, http.StatusOK, out)
}
