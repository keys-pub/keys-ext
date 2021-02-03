package server

import (
	"context"
	"net/http"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	wsapi "github.com/keys-pub/keys-ext/ws/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/encoding"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) postDirectMessage(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	body, st, err := readBody(c, true, 16*1024)
	if err != nil {
		return s.ErrResponse(c, st, err)
	}

	auth, _, err := s.auth(c, newAuth("Authorization", "sender", body))
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
		return s.ErrForbidden(c, errors.Errorf("not authorized to direct"))
	}

	path := dstore.Path("dms", recipient)
	_, idx, err := s.fi.EventsAdd(ctx, path, [][]byte{body})
	if err != nil {
		return err
	}

	dt, err := s.loadDirectToken(ctx, recipient)
	if err != nil {
		return s.ErrInternalServer(c, err)
	}
	if err := s.notifyDirectMessage(ctx, dt, idx); err != nil {
		return err
	}
	var out struct{}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) loadDirectToken(ctx context.Context, user keys.ID) (*api.DirectToken, error) {
	var dt api.DirectToken
	if _, err := s.fi.Load(ctx, dstore.Path("directs", user), &dt); err != nil {
		return nil, err
	}
	if dt.Token == "" {
		// Generate a direct token if one doesn't exist.
		dt.User = user
		dt.Token = encoding.MustEncode(keys.Rand32()[:], encoding.Base62)
		if err := s.fi.Set(ctx, dstore.Path("directs", user), dstore.From(dt), dstore.MergeAll()); err != nil {
			return nil, err
		}
	}
	return &dt, nil
}

func (s *Server) notifyDirectMessage(ctx context.Context, dt *api.DirectToken, idx int64) error {
	if s.internalKey == nil {
		return errors.Errorf("no secret key set")
	}
	event := &wsapi.Event{
		User:  dt.User,
		Index: idx,
		Token: dt.Token,
	}
	b, err := wsapi.Encrypt(event, s.internalKey)
	if err != nil {
		return err
	}
	if err := s.rds.Publish(ctx, wsapi.EventPubSub, b); err != nil {
		return err
	}
	return nil
}

func (s *Server) getDirectMessages(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	// ctx := c.Request().Context()

	auth, _, err := s.auth(c, newAuth("Authorization", "recipient", nil))
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

	out := &api.Events{
		Events:    resp.Events,
		Index:     resp.Index,
		Truncated: truncated,
	}

	return JSON(c, http.StatusOK, out)
}

func (s *Server) getDirectToken(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	auth, _, err := s.auth(c, newAuth("Authorization", "recipient", nil))
	if err != nil {
		return s.ErrForbidden(c, err)
	}

	dt, err := s.loadDirectToken(ctx, auth.KID)
	if err != nil {
		return s.ErrInternalServer(c, err)
	}

	out := &api.DirectToken{
		User:  dt.User,
		Token: dt.Token,
	}
	return JSON(c, http.StatusOK, out)
}
