package server

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) authChannel(c echo.Context, param string, content []byte) (*http.AuthResult, *http.AuthResult, error) {
	auth, err := s.auth(c, newAuth("Authorization", "", content))
	if err != nil {
		return nil, nil, err
	}

	// Skip nonce check here since the previous auth checks it.
	channel, err := s.auth(c, newAuth("Authorization-Channel", param, content).skipNonceCheck())
	if err != nil {
		return nil, nil, err
	}

	ctx := c.Request().Context()
	isMember, err := s.isChannelMember(ctx, channel.KID, auth.KID)
	if err != nil {
		return nil, nil, err
	}
	if !isMember {
		return nil, nil, errors.Errorf("not a member")
	}
	return channel, auth, nil
}

func (s *Server) putChannel(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	// We don't use authChannel here because the channel doesn't exist yet.
	auth, err := s.auth(c, newAuth("Authorization", "", nil))
	if err != nil {
		return ErrForbidden(c, err)
	}
	// Skip nonce check here since the previous auth checks it.
	channel, err := s.auth(c, newAuth("Authorization-Channel", "cid", nil).skipNonceCheck())
	if err != nil {
		return ErrForbidden(c, err)
	}

	ctx := c.Request().Context()
	path := dstore.Path("channels", channel.KID)

	if err := s.fi.Create(ctx, path, map[string]interface{}{"kid": channel.KID}); err != nil {
		switch err.(type) {
		case dstore.ErrPathExists:
			return ErrConflict(c, errors.Errorf("channel already exists"))
		}
		return s.internalError(c, err)
	}
	member := &api.ChannelMember{
		KID:  auth.KID,
		CID:  channel.KID,
		From: auth.KID,
	}
	if err := s.addChannelMembers(ctx, channel.KID, auth.KID, member); err != nil {
		return s.internalError(c, err)
	}

	var out struct{}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) isChannelMember(ctx context.Context, cid keys.ID, sid keys.ID) (bool, error) {
	// TODO: Cache this?
	path := dstore.Path("channels", cid, "members", sid)
	return s.fi.Exists(ctx, path)
}

func (s *Server) addChannelMembers(ctx context.Context, cid keys.ID, from keys.ID, members ...*api.ChannelMember) error {
	for _, member := range members {
		path := dstore.Path("channels", cid, "members", member.KID)
		add := &api.ChannelMember{
			KID:  member.KID,
			CID:  cid,
			From: from,
		}
		if err := s.fi.Create(ctx, path, dstore.From(add)); err != nil {
			return err
		}
		inboxPath := dstore.Path("inbox", member.KID, "channels", cid)
		if err := s.fi.Create(ctx, inboxPath, dstore.From(member)); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) getChannelMembers(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	channel, _, err := s.authChannel(c, "cid", nil)
	if err != nil {
		return ErrForbidden(c, err)
	}

	path := dstore.Path("channels", channel.KID, "members")
	iter, err := s.fi.DocumentIterator(ctx, path)
	if err != nil {
		return s.internalError(c, err)
	}
	defer iter.Release()
	members := []*api.ChannelMember{}
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
		members = append(members, &member)
	}
	out := &api.ChannelMembersResponse{
		Members: members,
	}
	return c.JSON(http.StatusOK, out)
}

func (s *Server) putChannelInfo(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	if c.Request().Body == nil {
		return ErrBadRequest(c, errors.Errorf("missing body"))
	}

	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return s.internalError(c, err)
	}

	if len(b) > 16*1024 {
		// TODO: Check length before reading data
		return ErrBadRequest(c, errors.Errorf("channel info too large (greater than 16KiB)"))
	}

	channel, _, err := s.authChannel(c, "cid", b)
	if err != nil {
		return ErrForbidden(c, err)
	}

	ctx := c.Request().Context()
	path := dstore.Path("channels", channel.KID)

	if err := s.fi.Set(ctx, path, map[string]interface{}{"info": b}, dstore.MergeAll()); err != nil {
		return s.internalError(c, err)
	}

	var out struct{}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) getChannelInfo(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	channel, _, err := s.authChannel(c, "cid", nil)
	if err != nil {
		return ErrForbidden(c, err)
	}

	path := dstore.Path("channels", channel.KID)
	doc, err := s.fi.Get(ctx, path)
	if err != nil {
		return s.internalError(c, err)
	}
	if doc == nil {
		return ErrNotFound(c, errors.Errorf("info not set"))
	}
	b := doc.Bytes("info")
	if b == nil {
		return ErrNotFound(c, errors.Errorf("info not set"))
	}
	return c.Blob(http.StatusOK, echo.MIMEOctetStream, b)
}

func (s *Server) postChannelInvite(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	if c.Request().Body == nil {
		return ErrBadRequest(c, errors.Errorf("missing body"))
	}

	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return s.internalError(c, err)
	}

	channel, auth, err := s.authChannel(c, "cid", b)
	if err != nil {
		return ErrForbidden(c, err)
	}

	var invite api.ChannelInvite
	if err := json.Unmarshal(b, &invite); err != nil {
		return ErrBadRequest(c, errors.Errorf("invalid channel invite"))
	}

	if invite.CID != channel.KID {
		return ErrBadRequest(c, errors.Errorf("invalid channel invite kid"))
	}
	if invite.Sender != auth.KID {
		return ErrBadRequest(c, errors.Errorf("invalid channel invite sender"))
	}
	if len(invite.EncryptedKey) > 1024 {
		return ErrBadRequest(c, errors.Errorf("invalid channel invite key"))
	}
	rid, err := keys.ParseID(invite.Recipient.String())
	if err != nil {
		return ErrBadRequest(c, errors.Errorf("invalid channel recipient"))
	}

	// TODO: Ensure inbox invites aren't full (over some threshold)
	// TODO: Restrict invite from a user@service

	invitePath := dstore.Path("channels", channel.KID, "invites", rid)

	exists, err := s.fi.Exists(ctx, invitePath)
	if err != nil {
		return s.internalError(c, err)
	}
	if exists {
		return ErrConflict(c, errors.Errorf("invite already exists"))
	}

	val := dstore.From(invite)
	if err := s.fi.Set(ctx, invitePath, val); err != nil {
		return s.internalError(c, err)
	}

	inboxPath := dstore.Path("inbox", rid, "invites", invite.CID)
	if err := s.fi.Set(ctx, inboxPath, val); err != nil {
		return s.internalError(c, err)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) getChannelInvites(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	channel, _, err := s.authChannel(c, "cid", nil)
	if err != nil {
		return ErrForbidden(c, err)
	}

	path := dstore.Path("channels", channel.KID, "invites")
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

// func (s *Server) postChannelMembers(c echo.Context) error {
// 	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

// 	if c.Request().Body == nil {
// 		return ErrBadRequest(c, errors.Errorf("missing body"))
// 	}

// 	b, err := ioutil.ReadAll(c.Request().Body)
// 	if err != nil {
// 		return s.internalError(c, err)
// 	}

// 	if len(b) > 16*1024 {
// 		// TODO: Check length before reading data
// 		return ErrBadRequest(c, errors.Errorf("channel data too large (greater than 16KiB)"))
// 	}

// 	channel, auth, err := s.authChannel(c, "cid", b)
// 	if err != nil {
// 		s.logger.Errorf("Auth failed: %v", err)
// 		return ErrForbidden(c, err)
// 	}

// 	var req api.ChannelMembersAddRequest
// 	if err := json.Unmarshal(b, &req); err != nil {
// 		return ErrBadRequest(c, err)
// 	}

// 	ctx := c.Request().Context()
// 	if err := s.addChannelMembers(ctx, channel.KID, auth.KID, req.Members); err != nil {
// 		return s.internalError(c, err)
// 	}
// 	var out struct{}
// 	return JSON(c, http.StatusOK, out)
// }
