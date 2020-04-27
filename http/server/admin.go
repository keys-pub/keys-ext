package server

import (
	"net/http"

	"github.com/keys-pub/keys"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) adminCheck(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	request := c.Request()
	ctx := request.Context()

	auth, status, err := checkAuth(c, s.URL, s.nowFn(), s.mc)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}
	if !s.isAdmin(auth.KID) {
		return ErrForbidden(c, errors.Errorf("not authorized"))
	}

	if c.Param("kid") == "all" {
		kids, err := s.users.KIDs(ctx)
		if err != nil {
			return s.internalError(c, err)
		}

		for _, kid := range kids {
			if err := s.tasks.CreateTask(ctx, "POST", "/task/check/"+kid.String(), s.internalAuth); err != nil {
				return s.internalError(c, err)
			}
		}
	} else {
		kid, err := keys.ParseID(c.Param("kid"))
		if err != nil {
			return ErrNotFound(c, errors.Errorf("kid not found"))
		}
		if err := s.tasks.CreateTask(ctx, "POST", "/task/check/"+kid.String(), s.internalAuth); err != nil {
			return s.internalError(c, err)
		}
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) isAdmin(kid keys.ID) bool {
	for _, admin := range s.admins {
		if admin == kid {
			return true
		}
	}
	return false
}
