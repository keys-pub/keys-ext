package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) listUserChannels(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	auth, err := s.auth(c, newAuth("Authorization", "kid", nil))
	if err != nil {
		return ErrForbidden(c, err)
	}

	path := dstore.Path("users", auth.KID, "channels")
	iter, err := s.fi.DocumentIterator(ctx, path, dstore.NoData())
	if err != nil {
		return s.internalError(c, err)
	}
	defer iter.Release()
	paths := []string{}
	for {
		doc, err := iter.Next()
		if err != nil {
			return s.internalError(c, err)
		}
		if doc == nil {
			break
		}
		paths = append(paths, dstore.Path("channels", dstore.PathLast(doc.Path)))
	}

	channels := make([]*api.Channel, 0, len(paths))

	positions, err := s.fi.EventPositions(ctx, paths)
	if err != nil {
		return s.internalError(c, err)
	}
	for _, pos := range positions {
		channels = append(channels, &api.Channel{
			ID:        keys.ID(dstore.PathLast(pos.Path)),
			Index:     pos.Index,
			Timestamp: pos.Timestamp,
		})
	}

	out := &api.UserChannelsResponse{Channels: channels}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) getUserChannelInvites(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	auth, err := s.auth(c, newAuth("Authorization", "kid", nil))
	if err != nil {
		return ErrForbidden(c, err)
	}

	path := dstore.Path("users", auth.KID, "invites")
	iter, err := s.fi.DocumentIterator(ctx, path)
	if err != nil {
		return s.internalError(c, err)
	}
	defer iter.Release()
	invites := []*api.ChannelInvite{}
	for {
		doc, err := iter.Next()
		if err != nil {
			return s.internalError(c, err)
		}
		if doc == nil {
			break
		}
		var invite api.ChannelInvite
		if err := doc.To(&invite); err != nil {
			return s.internalError(c, err)
		}
		invites = append(invites, &invite)
	}
	out := &api.ChannelInvitesResponse{Invites: invites}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) getUserChannelInvite(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	auth, err := s.auth(c, newAuth("Authorization", "kid", nil))
	if err != nil {
		return ErrForbidden(c, err)
	}

	cid, err := keys.ParseID(c.Param("cid"))
	if err != nil {
		return ErrBadRequest(c, err)
	}

	path := dstore.Path("users", auth.KID, "invites", cid)
	doc, err := s.fi.Get(ctx, path)
	if err != nil {
		return s.internalError(c, err)
	}
	if doc == nil {
		return ErrNotFound(c, errors.Errorf("invite not found"))
	}

	var invite api.ChannelInvite
	if err := doc.To(&invite); err != nil {
		return s.internalError(c, err)
	}

	out := &api.UserChannelInviteResponse{Invite: &invite}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) deleteUserChannelInvite(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	auth, err := s.auth(c, newAuth("Authorization", "kid", nil))
	if err != nil {
		return ErrForbidden(c, err)
	}

	cid, err := keys.ParseID(c.Param("cid"))
	if err != nil {
		return ErrBadRequest(c, err)
	}

	ok, err := s.deleteInvite(ctx, cid, auth.KID)
	if err != nil {
		return s.internalError(c, err)
	}
	if !ok {
		return ErrNotFound(c, errors.Errorf("invite not found"))
	}

	var out struct{}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) deleteInvite(ctx context.Context, cid keys.ID, kid keys.ID) (bool, error) {
	path := dstore.Path("users", kid, "invites", cid)
	ok, err := s.fi.Delete(ctx, path)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}

	invitePath := dstore.Path("channels", cid, "invites", kid)
	if _, err := s.fi.Delete(ctx, invitePath); err != nil {
		return false, err
	}

	return true, nil
}

// putUserChannel to join a channel.
// Authorization and Authorization-Channel required.
func (s *Server) putUserChannel(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	body, st, err := readBody(c, false, 64*1024)
	if err != nil {
		return ErrResponse(c, st, err)
	}

	auth, err := s.auth(c, newAuth("Authorization", "kid", body))
	if err != nil {
		return ErrForbidden(c, err)
	}

	// Skip nonce check here since the previous auth checks it.
	channel, err := s.auth(c, newAuth("Authorization-Channel", "cid", body).skipNonceCheck())
	if err != nil {
		return ErrForbidden(c, err)
	}

	var req api.ChannelJoinRequest
	if len(body) != 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			return ErrBadRequest(c, errors.Errorf("invalid channel join request"))
		}
	}

	path := dstore.Path("users", auth.KID, "invites", channel.KID)
	doc, err := s.fi.Get(ctx, path)
	if err != nil {
		return s.internalError(c, err)
	}
	if doc == nil {
		// User had channel key, but no current invite.
		return ErrNotFound(c, errors.Errorf("invite not found"))
	}
	var invite api.ChannelInvite
	if err := doc.To(&invite); err != nil {
		return s.internalError(c, err)
	}

	if err := s.addChannelUsers(ctx, channel.KID, auth.KID); err != nil {
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

func (s *Server) deleteUserChannel(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	body, st, err := readBody(c, false, 64*1024)
	if err != nil {
		return ErrResponse(c, st, err)
	}

	auth, err := s.auth(c, newAuth("Authorization", "kid", body))
	if err != nil {
		return ErrForbidden(c, err)
	}
	cid, err := keys.ParseID(c.Param("cid"))
	if err != nil {
		return ErrBadRequest(c, err)
	}

	var req api.ChannelLeaveRequest
	if len(body) != 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			return ErrBadRequest(c, errors.Errorf("invalid channel leave request"))
		}
	}

	if len(req.Message) > 0 {
		if err := s.sendMessage(c, cid, req.Message); err != nil {
			return s.internalError(c, err)
		}
	}

	if err := s.removeChannelUsers(ctx, cid, auth.KID); err != nil {
		return s.internalError(c, err)
	}

	var out struct{}
	return JSON(c, http.StatusOK, out)
}
