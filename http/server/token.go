package server

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
	"github.com/labstack/echo/v4"
)

func (s *Server) postUserToken(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	auth, err := s.auth(c, newAuth("Authorization", "kid", nil))
	if err != nil {
		return ErrForbidden(c, err)
	}

	token := api.GenerateToken()

	if err := s.fi.Set(ctx, dstore.Path("users", auth.KID, "tokens", token), dstore.Empty()); err != nil {
		return s.internalError(c, err)
	}

	out := &api.UserTokenResponse{Token: token}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) checkUserToken(ctx context.Context, token string, user keys.ID) (bool, error) {
	// TODO: Validate user token format
	if token == "" {
		return false, nil
	}
	return s.fi.Exists(ctx, dstore.Path("users", user, "tokens", token))
}
