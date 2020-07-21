package server

import (
	"context"
	"net/http"

	"github.com/keys-pub/keys"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) check(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	request := c.Request()
	ctx := request.Context()

	// Auth
	auth := request.Header.Get("Authorization")
	if auth == "" {
		return ErrUnauthorized(c, errors.Errorf("missing Authorization header"))
	}
	res, status, err := checkAuth(c, s.URL, "", s.clock.Now(), s.rds)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	kid := res.KID
	if err := s.checkKID(ctx, kid); err != nil {
		return s.internalError(c, err)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) checkKID(ctx context.Context, kid keys.ID) error {
	return s.tasks.CreateTask(ctx, "POST", "/task/check/"+kid.String(), s.internalAuth)
}
