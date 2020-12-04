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

	body, st, err := readBody(c, true, 64*1024)
	if err != nil {
		return ErrResponse(c, st, err)
	}

	channel, _, err := s.authChannel(c, "cid", body)
	if err != nil {
		return ErrForbidden(c, err)
	}

	if err := s.sendMessage(c, channel.KID, body); err != nil {
		return s.internalError(c, err)
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

	events, idx, err := s.fi.EventsAdd(ctx, path, [][]byte{msg})
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return errors.Errorf("no events added")
	}
	if err := s.notifyChannelMessage(ctx, channel, idx); err != nil {
		return err
	}
	return nil
}

func (s *Server) notifyChannelMessage(ctx context.Context, channel keys.ID, idx int64) error {
	if s.secretKey == nil {
		return errors.Errorf("no secret key set")
	}
	recipients, err := s.channelUserIDs(ctx, channel)
	if err != nil {
		return err
	}
	pub := &wsapi.PubSubEvent{
		Type:       wsapi.ChannelMessageEventType,
		Channel:    channel,
		Recipients: recipients,
		Index:      idx,
	}
	pb, err := wsapi.Encrypt(pub, s.secretKey)
	if err != nil {
		return err
	}
	if err := s.rds.Publish(ctx, wsapi.EventPubSub, pb); err != nil {
		return err
	}
	return nil
}

func (s *Server) listMessages(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	channel, _, err := s.authChannel(c, "cid", nil)
	if err != nil {
		return ErrForbidden(c, err)
	}

	limit := 1000
	path := dstore.Path("channels", channel.KID)
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

	return JSON(c, http.StatusOK, out)
}
