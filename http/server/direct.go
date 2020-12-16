package server

import (
	"net/http"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) postDirect(c echo.Context) error {
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

	token := c.Param("token")
	if token == "" {
	}
	hasToken, err := s.checkUserToken(ctx, token, kid)
	if err != nil {
		return s.internalError(c, err)
	}
	if !hasToken {
		return ErrForbidden(c, errors.Errorf("invalid token"))
	}
	path := dstore.Path("dms", kid)
	if _, _, err := s.fi.EventsAdd(ctx, path, [][]byte{body}); err != nil {
		return err
	}
	var out struct{}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) getDirects(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	auth, err := s.auth(c, newAuth("Authorization", "kid", nil))
	if err != nil {
		return ErrForbidden(c, err)
	}

	limit := 1000
	path := dstore.Path("dms", auth.KID)
	resp, err := s.events(c, path, limit)
	if err != nil {
		return s.internalError(c, err)
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
}
