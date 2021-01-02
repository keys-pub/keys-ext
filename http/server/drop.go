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
		return s.ErrResponse(c, st, err)
	}

	auth, err := s.auth(c, newAuth("Authorization", "sender", body))
	if err != nil {
		return s.ErrForbidden(c, err)
	}

	recipient, err := keys.ParseID(c.Param("recipient"))
	if err != nil {
		return s.ErrBadRequest(c, errors.Errorf("invalid recipient"))
	}

	// Does the recipient follow the sender?
	follow, err := s.follow(ctx, recipient, auth.KID)
	if err != nil {
		return s.ErrInternalServer(c, err)
	}
	if follow == nil {
		return s.ErrForbidden(c, errors.Errorf("not authorized to drop"))
	}

	path := dstore.Path("dms", recipient)
	if _, _, err := s.fi.EventsAdd(ctx, path, [][]byte{body}); err != nil {
		return err
	}
	var out struct{}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) getDrops(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	auth, err := s.auth(c, newAuth("Authorization", "recipient", nil))
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
