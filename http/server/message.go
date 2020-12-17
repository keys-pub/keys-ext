package server

import (
	"context"
	"encoding/json"
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

	body, st, err := readBody(c, true, 64*1024)
	if err != nil {
		return ErrResponse(c, st, err)
	}

	channel, err := s.auth(c, newAuth("Authorization", "cid", body))
	if err != nil {
		return ErrForbidden(c, err)
	}

	path := dstore.Path("channels", channel.KID)
	doc, err := s.fi.Get(ctx, path)
	if err != nil {
		return ErrInternalServer(c, err)
	}
	if doc == nil {
		return ErrNotFound(c, keys.NewErrNotFound(channel.KID.String()))
	}

	if err := s.sendMessage(c, channel.KID, body); err != nil {
		return ErrInternalServer(c, err)
	}

	var out struct{}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) sendMessage(c echo.Context, channel keys.ID, msg []byte) error {
	if len(msg) == 0 {
		return errors.Errorf("no message data")
	}
	ctx := c.Request().Context()
	path := dstore.Path("channels", channel)

	_, idx, err := s.fi.EventsAdd(ctx, path, [][]byte{msg})
	if err != nil {
		return err
	}
	if err := s.notifyChannelMessage(ctx, channel, idx); err != nil {
		return err
	}
	return nil
}

func (s *Server) notifyChannelMessage(ctx context.Context, channel keys.ID, idx int64) error {
	event := &wsapi.Event{
		Channel: channel,
		Index:   idx,
	}
	b, err := json.Marshal(event)
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

	channel, err := s.auth(c, newAuth("Authorization", "cid", nil))
	if err != nil {
		return ErrForbidden(c, err)
	}

	path := dstore.Path("channels", channel.KID)
	doc, err := s.fi.Get(ctx, path)
	if err != nil {
		return ErrInternalServer(c, err)
	}
	if doc == nil {
		return ErrNotFound(c, keys.NewErrNotFound(channel.KID.String()))
	}

	limit := 1000
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
