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

	auth, err := s.auth(c, newAuth("Authorization", "from", body))
	if err != nil {
		return ErrForbidden(c, err)
	}
	from := auth.KID

	to, err := keys.ParseID(c.Param("to"))
	if err != nil {
		return ErrBadRequest(c, errors.Errorf("invalid sender"))
	}

	following, err := s.follows(ctx, to, from)
	if err != nil {
		return ErrInternalServer(c, err)
	}
	if !following {
		return ErrForbidden(c, nil)
	}
	path := dstore.Path("dms", to)
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
	resp, st, err := s.events(c, path, limit)
	if err != nil {
		return ErrResponse(c, st, err)
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
