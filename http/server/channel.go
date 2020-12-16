package server

import (
	"context"
	"encoding/json"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	wsapi "github.com/keys-pub/keys-ext/ws/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
	"github.com/keys-pub/keys/tsutil"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

func (s *Server) putChannel(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	body, st, err := readBody(c, false, 64*1024)
	if err != nil {
		return ErrResponse(c, st, err)
	}

	channel, err := s.auth(c, newAuth("Authorization", "cid", body))
	if err != nil {
		return ErrForbidden(c, err)
	}

	var req api.ChannelCreateRequest
	if len(body) != 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			return ErrBadRequest(c, errors.Errorf("invalid channel create request"))
		}
	}

	ctx := c.Request().Context()
	path := dstore.Path("channels", channel.KID)

	create := &api.Channel{
		ID: channel.KID,
	}

	if err := s.fi.Create(ctx, path, dstore.From(create)); err != nil {
		switch err.(type) {
		case dstore.ErrPathExists:
			return ErrConflict(c, errors.Errorf("channel already exists"))
		}
		return s.internalError(c, err)
	}

	if err := s.notifyChannelCreated(ctx, channel.KID); err != nil {
		return s.internalError(c, err)
	}

	if len(req.Message) > 0 {
		if err := s.sendMessage(c, channel.KID, req.Message); err != nil {
			return s.internalError(c, err)
		}
	}

	var out struct{}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) notifyChannelCreated(ctx context.Context, channel keys.ID) error {
	pub := &wsapi.Event{
		Channel: channel,
	}
	b, err := msgpack.Marshal(pub)
	if err != nil {
		return err
	}
	if err := s.rds.Publish(ctx, wsapi.EventPubSub, b); err != nil {
		return err
	}
	return nil
}

func (s *Server) getChannel(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	channel, err := s.auth(c, newAuth("Authorization", "cid", nil))
	if err != nil {
		return ErrForbidden(c, err)
	}

	path := dstore.Path("channels", channel.KID)

	doc, err := s.fi.Get(ctx, path)
	if err != nil {
		return s.internalError(c, err)
	}
	if doc == nil {
		return ErrNotFound(c, keys.NewErrNotFound(channel.KID.String()))
	}

	var out api.Channel
	if err := doc.To(&out); err != nil {
		return s.internalError(c, err)
	}
	out.Timestamp = tsutil.Millis(doc.UpdatedAt)

	positions, err := s.fi.EventPositions(ctx, []string{path})
	if err != nil {
		return s.internalError(c, err)
	}
	if len(positions) > 0 {
		out.Index = positions[0].Index
		if positions[0].Timestamp > 0 {
			out.Timestamp = positions[0].Timestamp
		}
	}
	return c.JSON(http.StatusOK, out)
}
