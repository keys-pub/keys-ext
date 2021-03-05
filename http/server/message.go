package server

import (
	"context"
	"net/http"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	wsapi "github.com/keys-pub/keys-ext/ws/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) postMessage(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	body, err := readBody(c, true, 64*1024)
	if err != nil {
		return s.ErrResponse(c, err)
	}

	auth, _, err := s.auth(c, newAuth("Authorization", "cid", body))
	if err != nil {
		return s.ErrForbidden(c, err)
	}

	path := dstore.Path("channels", auth.KID)
	doc, err := s.fi.Get(ctx, path)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	if doc == nil {
		return s.ErrNotFound(c, keys.NewErrNotFound(auth.KID.String()))
	}
	var channel api.Channel
	if err := doc.To(&channel); err != nil {
		return s.ErrResponse(c, err)
	}

	ct := &api.ChannelToken{Channel: channel.ID, Token: channel.Token}
	if err := s.sendMessage(c, ct, body); err != nil {
		return s.ErrResponse(c, err)
	}

	var out struct{}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) sendMessage(c echo.Context, ct *api.ChannelToken, msg []byte) error {
	if len(msg) == 0 {
		return errors.Errorf("empty message data")
	}
	if ct.Token == "" {
		return errors.Errorf("empty token")
	}
	ctx := c.Request().Context()
	path := dstore.Path("channels", ct.Channel)

	_, idx, err := s.fi.EventsAdd(ctx, path, [][]byte{msg})
	if err != nil {
		return err
	}
	if err := s.notifyChannelMessage(ctx, ct, idx); err != nil {
		return err
	}
	return nil
}

func (s *Server) notifyChannelMessage(ctx context.Context, ct *api.ChannelToken, idx int64) error {
	if s.internalKey == nil {
		return errors.Errorf("no secret key set")
	}
	event := &wsapi.Event{
		Channel: ct.Channel,
		Index:   idx,
		Token:   ct.Token,
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

func (s *Server) getMessages(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	auth, _, err := s.auth(c, newAuth("Authorization", "cid", nil))
	if err != nil {
		return s.ErrForbidden(c, err)
	}

	path := dstore.Path("channels", auth.KID)
	doc, err := s.fi.Get(ctx, path)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	if doc == nil {
		return s.ErrNotFound(c, keys.NewErrNotFound(auth.KID.String()))
	}
	var channel api.Channel
	if err := doc.To(&channel); err != nil {
		return s.ErrResponse(c, err)
	}

	limit := 1000
	resp, err := s.events(c, path, limit)
	if err != nil {
		return s.ErrResponse(c, err)
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
