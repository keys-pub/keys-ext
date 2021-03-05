package server

import (
	"context"
	"net/http"

	"github.com/keys-pub/keys"
	"github.com/labstack/echo/v4"
)

func (s *Server) check(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	request := c.Request()
	ctx := request.Context()

	auth, _, err := s.auth(c, newAuth("Authorization", "", nil))
	if err != nil {
		return s.ErrForbidden(c, err)
	}

	if err := s.checkKID(ctx, auth.KID, HighPriority); err != nil {
		return s.ErrResponse(c, err)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) checkKID(ctx context.Context, kid keys.ID, priority TaskPriority) error {
	return s.tasks.CreateTask(ctx, "POST", "/task/check/"+kid.String(), s.internalAuth, priority)
}
