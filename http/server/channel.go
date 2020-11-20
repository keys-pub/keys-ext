package server

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
	"github.com/keys-pub/keys/tsutil"
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

	create := &api.Channel{
		ID:      channel.KID,
		Creator: auth.KID,
	}

	if err := s.fi.Create(ctx, path, dstore.From(create)); err != nil {
		switch err.(type) {
		case dstore.ErrPathExists:
			return ErrConflict(c, errors.Errorf("channel already exists"))
		}
		return s.internalError(c, err)
	}
	user := &api.ChannelUser{
		User:    auth.KID,
		Channel: channel.KID,
		From:    auth.KID,
	}
	if err := s.addChannelUsers(ctx, channel.KID, auth.KID, user); err != nil {
		return s.internalError(c, err)
	}

	var out struct{}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) getChannel(c echo.Context) error {
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

func (s *Server) isChannelMember(ctx context.Context, channel keys.ID, user keys.ID) (bool, error) {
	// TODO: Cache this?
	path := dstore.Path("channels", channel, "users", user)
	return s.fi.Exists(ctx, path)
}

func (s *Server) channelUserIDs(ctx context.Context, channel keys.ID) ([]keys.ID, error) {
	path := dstore.Path("channels", channel, "users")
	iter, err := s.fi.DocumentIterator(ctx, path, dstore.NoData())
	if err != nil {
		return nil, err
	}
	defer iter.Release()
	kids := []keys.ID{}
	for {
		doc, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if doc == nil {
			break
		}
		kids = append(kids, keys.ID(dstore.PathLast(doc.Path)))
	}
	return kids, nil
}

func (s *Server) addChannelUsers(ctx context.Context, channel keys.ID, from keys.ID, users ...*api.ChannelUser) error {
	// TODO: Before adding check if limits on number of channels for user
	for _, user := range users {
		path := dstore.Path("channels", channel, "users", user.User)
		if user.Channel != channel {
			return errors.Errorf("user channel mismatch")
		}
		add := &api.ChannelUser{
			Channel: channel,
			User:    user.User,
			From:    from,
		}
		if err := s.fi.Create(ctx, path, dstore.From(add)); err != nil {
			return err
		}
		usersPath := dstore.Path("users", user.User, "channels", channel)
		ch := &api.Channel{
			ID: channel,
		}
		if err := s.fi.Create(ctx, usersPath, dstore.From(ch)); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) getChannelUsers(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	channel, _, err := s.authChannel(c, "cid", nil)
	if err != nil {
		return ErrForbidden(c, err)
	}

	path := dstore.Path("channels", channel.KID, "users")
	iter, err := s.fi.DocumentIterator(ctx, path)
	if err != nil {
		return s.internalError(c, err)
	}
	defer iter.Release()
	users := []*api.ChannelUser{}
	for {
		doc, err := iter.Next()
		if err != nil {
			return s.internalError(c, err)
		}
		if doc == nil {
			break
		}
		var user api.ChannelUser
		if err := doc.To(&user); err != nil {
			return s.internalError(c, err)
		}
		users = append(users, &user)
	}
	out := &api.ChannelUsersResponse{
		Users: users,
	}
	return c.JSON(http.StatusOK, out)
}

func (s *Server) postChannelInvites(c echo.Context) error {
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

	var invites []*api.ChannelInvite
	if err := json.Unmarshal(b, &invites); err != nil {
		return ErrBadRequest(c, errors.Errorf("invalid channel invites"))
	}

	if len(invites) > 10 {
		return ErrBadRequest(c, errors.Errorf("too many invites"))
	}

	for _, invite := range invites {
		if invite.Channel != channel.KID {
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

		// TODO: Ensure users invites aren't full (over some threshold)
		// TODO: Restrict invite from a user@service

		invitePath := dstore.Path("channels", channel.KID, "invites", rid)

		// TODO: If they have an existing invite it will get overwritten,
		// with the same data, although maybe from a different sender.
		// exists, err := s.fi.Exists(ctx, invitePath)
		// if err != nil {
		// 	return s.internalError(c, err)
		// }
		// if exists {
		// 	return ErrConflict(c, errors.Errorf("invite already exists"))
		// }

		val := dstore.From(invite)
		if err := s.fi.Set(ctx, invitePath, val); err != nil {
			return s.internalError(c, err)
		}

		usersPath := dstore.Path("users", rid, "invites", invite.Channel)
		if err := s.fi.Set(ctx, usersPath, val); err != nil {
			return s.internalError(c, err)
		}
	}

	var out struct{}
	return JSON(c, http.StatusOK, out)
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

// func (s *Server) postChannelUsers(c echo.Context) error {
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

// 	var req api.ChannelUsersAddRequest
// 	if err := json.Unmarshal(b, &req); err != nil {
// 		return ErrBadRequest(c, err)
// 	}

// 	ctx := c.Request().Context()
// 	if err := s.addChannelUsers(ctx, channel.KID, auth.KID, req.Users); err != nil {
// 		return s.internalError(c, err)
// 	}
// 	var out struct{}
// 	return JSON(c, http.StatusOK, out)
// }
