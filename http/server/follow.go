package server

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) postFollow(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	auth, err := s.auth(c, newAuth("Authorization", "kid", nil))
	if err != nil {
		return ErrForbidden(c, err)
	}

	user, err := keys.ParseID(c.Param("user"))
	if err != nil {
		return ErrBadRequest(c, errors.Errorf("invalid user"))
	}

	follow := &api.Follow{KID: auth.KID, User: user}
	if err := s.fi.Set(ctx, dstore.Path("follows", auth.KID, "users", user), dstore.From(follow)); err != nil {
		return ErrInternalServer(c, err)
	}

	var out struct{}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) follows(ctx context.Context, kid keys.ID, user keys.ID) (bool, error) {
	return s.fi.Exists(ctx, dstore.Path("follows", kid, "users", user))
}

func (s *Server) getFollows(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	auth, err := s.auth(c, newAuth("Authorization", "kid", nil))
	if err != nil {
		return ErrForbidden(c, err)
	}

	iter, err := s.fi.DocumentIterator(ctx, dstore.Path("follows", auth.KID, "users"))
	if err != nil {
		return ErrInternalServer(c, err)
	}
	follows := []*api.Follow{}
	for {
		doc, err := iter.Next()
		if err != nil {
			return ErrInternalServer(c, err)
		}
		if doc == nil {
			break
		}
		var follow api.Follow
		if err := doc.To(&follow); err != nil {
			return ErrInternalServer(c, err)
		}
		follows = append(follows, &follow)
	}

	out := api.FollowsResponse{Follows: follows}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) deleteFollow(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	auth, err := s.auth(c, newAuth("Authorization", "kid", nil))
	if err != nil {
		return ErrForbidden(c, err)
	}

	user, err := keys.ParseID(c.Param("user"))
	if err != nil {
		return ErrBadRequest(c, errors.Errorf("invalid user"))
	}

	ok, err := s.fi.Delete(ctx, dstore.Path("follows", auth.KID, "users", user))
	if err != nil {
		return ErrInternalServer(c, err)
	}
	if !ok {
		return ErrNotFound(c, errors.Errorf("follow not found"))
	}

	var out struct{}
	return JSON(c, http.StatusOK, out)
}
