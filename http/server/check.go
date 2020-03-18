package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) check(c echo.Context) error {
	request := c.Request()
	ctx := request.Context()
	logger.Infof(ctx, "Server GET check %s", s.urlString(c))

	// Auth
	auth := request.Header.Get("Authorization")
	if auth == "" {
		return ErrUnauthorized(c, errors.Errorf("missing Authorization header"))
	}
	now := s.nowFn()
	authRes, err := CheckAuthorization(request.Context(), request.Method, s.urlString(c), auth, s.mc, now)
	if err != nil {
		return ErrForbidden(c, err)
	}
	kid := authRes.kid

	if err := s.tasks.CreateTask(ctx, "POST", "/task/check/"+kid.String(), s.internalAuth); err != nil {
		return internalError(c, err)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}
