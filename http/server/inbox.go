package server

import (
	"net/http"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) inboxChannels(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	auth, err := s.auth(c, newAuth("Authorization", "kid", nil))
	if err != nil {
		return ErrForbidden(c, err)
	}

	path := dstore.Path("inbox", auth.KID, "channels")
	iter, err := s.fi.DocumentIterator(ctx, path)
	if err != nil {
		return s.internalError(c, err)
	}
	defer iter.Release()
	channels := []*api.Channel{}
	for {
		doc, err := iter.Next()
		if err != nil {
			return s.internalError(c, err)
		}
		if doc == nil {
			break
		}
		var member api.ChannelMember
		if err := doc.To(&member); err != nil {
			return s.internalError(c, err)
		}
		channels = append(channels, &api.Channel{
			ID: member.CID,
		})
	}
	resp := &api.InboxChannelsResponse{Channels: channels}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) inboxInvites(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	auth, err := s.auth(c, newAuth("Authorization", "kid", nil))
	if err != nil {
		return ErrForbidden(c, err)
	}

	path := dstore.Path("inbox", auth.KID, "invites")
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
	resp := &api.ChannelInvitesResponse{Invites: invites}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) deleteInboxInvite(c echo.Context) error {
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

	path := dstore.Path("inbox", auth.KID, "invites", cid)
	ok, err := s.fi.Delete(ctx, path)
	if err != nil {
		return s.internalError(c, err)
	}
	if !ok {
		return ErrNotFound(c, errors.Errorf("invite not found"))
	}
	var out struct{}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) acceptInboxInvite(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	auth, err := s.auth(c, newAuth("Authorization", "kid", nil))
	if err != nil {
		return ErrForbidden(c, err)
	}

	// Skip nonce check here since the previous auth checks it.
	channel, err := s.auth(c, newAuth("Authorization-Channel", "cid", nil).skipNonceCheck())
	if err != nil {
		return ErrForbidden(c, err)
	}

	path := dstore.Path("inbox", auth.KID, "invites", channel.KID)
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

	member := &api.ChannelMember{
		KID:  auth.KID,
		CID:  channel.KID,
		From: invite.Sender,
	}
	if err := s.addChannelMembers(ctx, channel.KID, auth.KID, member); err != nil {
		return s.internalError(c, err)
	}
	var out struct{}
	return JSON(c, http.StatusOK, out)
}
