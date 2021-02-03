package server

import (
	"encoding/json"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keys/http"
	"github.com/keys-pub/keys/tsutil"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) putChannel(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	body, st, err := readBody(c, false, 64*1024)
	if err != nil {
		return s.ErrResponse(c, st, err)
	}

	channel, _, err := s.auth(c, newAuth("Authorization", "cid", body))
	if err != nil {
		return s.ErrForbidden(c, err)
	}

	var req api.ChannelCreateRequest
	if len(body) != 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			return s.ErrBadRequest(c, errors.Errorf("invalid channel create request"))
		}
	}

	ctx := c.Request().Context()
	path := dstore.Path("channels", channel.KID)

	token := encoding.MustEncode(keys.RandBytes(32), encoding.Base62)
	create := &api.Channel{
		ID:    channel.KID,
		Token: token,
	}

	if err := s.fi.Create(ctx, path, dstore.From(create)); err != nil {
		switch err.(type) {
		case dstore.ErrPathExists:
			return s.ErrConflict(c, errors.Errorf("channel already exists"))
		}
		return s.ErrInternalServer(c, err)
	}

	if len(req.Message) > 0 {
		ct := &api.ChannelToken{
			Channel: channel.KID,
			Token:   token,
		}
		if err := s.sendMessage(c, ct, req.Message); err != nil {
			return s.ErrInternalServer(c, err)
		}
	}

	out := &api.ChannelCreateResponse{
		Channel: create,
	}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) getChannel(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	channel, _, err := s.auth(c, newAuth("Authorization", "cid", nil))
	if err != nil {
		return s.ErrForbidden(c, err)
	}

	path := dstore.Path("channels", channel.KID)

	doc, err := s.fi.Get(ctx, path)
	if err != nil {
		return s.ErrInternalServer(c, err)
	}
	if doc == nil {
		return s.ErrNotFound(c, keys.NewErrNotFound(channel.KID.String()))
	}

	var out api.Channel
	if err := doc.To(&out); err != nil {
		return s.ErrInternalServer(c, err)
	}
	out.Timestamp = tsutil.Millis(doc.UpdatedAt)

	positions, err := s.fi.EventPositions(ctx, []string{path})
	if err != nil {
		return s.ErrInternalServer(c, err)
	}
	position := positions[path]
	if position != nil {
		out.Index = position.Index
		if position.Timestamp > 0 {
			out.Timestamp = position.Timestamp
		}
	}
	return c.JSON(http.StatusOK, out)
}

func (s *Server) postChannelsStatus(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	body, st, err := readBody(c, false, 64*1024)
	if err != nil {
		return s.ErrResponse(c, st, err)
	}
	var req api.ChannelsStatusRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return s.ErrBadRequest(c, errors.Errorf("invalid request"))
	}

	paths := []string{}
	for cid := range req.Channels {
		channel, err := keys.ParseID(string(cid))
		if err != nil {
			return s.ErrBadRequest(c, errors.Errorf("invalid request"))
		}
		paths = append(paths, dstore.Path("channels", channel))
	}

	docs, err := s.fi.GetAll(ctx, paths)
	if err != nil {
		return s.ErrInternalServer(c, err)
	}
	positions, err := s.fi.EventPositions(ctx, paths)
	if err != nil {
		return s.ErrInternalServer(c, err)
	}

	channels := make([]*api.ChannelStatus, 0, len(docs))
	for _, doc := range docs {
		var channel api.Channel
		if err := doc.To(&channel); err != nil {
			return s.ErrInternalServer(c, err)
		}
		token := req.Channels[channel.ID]
		if token == "" || token != channel.Token {
			continue
		}
		channel.Timestamp = tsutil.Millis(doc.UpdatedAt)
		position := positions[doc.Path]
		if position != nil {
			channel.Index = position.Index
			if position.Timestamp > 0 {
				channel.Timestamp = position.Timestamp
			}
		}
		channels = append(channels, &api.ChannelStatus{
			ID:        channel.ID,
			Index:     channel.Index,
			Timestamp: channel.Timestamp,
		})
	}

	out := api.ChannelsStatusResponse{
		Channels: channels,
	}
	return c.JSON(http.StatusOK, out)
}
