package server

import (
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/http"
	"github.com/keys-pub/keys/user"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) adminCheck(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	request := c.Request()
	ctx := request.Context()

	auth, err := s.auth(c, newAuth("Authorization", "", nil))
	if err != nil {
		return ErrForbidden(c, err)
	}

	if !s.isAdmin(auth.KID) {
		return ErrForbidden(c, errors.Errorf("not authorized"))
	}

	switch c.Param("kid") {
	case "all":
		kids, err := s.users.KIDs(ctx)
		if err != nil {
			return ErrInternalServer(c, err)
		}
		s.logger.Infof("Checking all (%d)", len(kids))
		if err := s.checkKeys(ctx, kids); err != nil {
			return ErrInternalServer(c, err)
		}
	case "content-not-found":
		if err := s.checkUserStatus(ctx, user.StatusContentNotFound); err != nil {
			return ErrInternalServer(c, err)
		}
	case "connection-fail":
		if err := s.checkUserStatus(ctx, user.StatusConnFailure); err != nil {
			return ErrInternalServer(c, err)
		}
	case "expired":
		if err := s.checkExpired(ctx, time.Hour*6, time.Hour*24*7); err != nil {
			return ErrInternalServer(c, err)
		}
	default:
		kid, err := keys.ParseID(c.Param("kid"))
		if err != nil {
			return ErrNotFound(c, errors.Errorf("invalid kid"))
		}
		s.logger.Infof("Checking %s", kid)
		if err := s.checkKID(ctx, kid, HighPriority); err != nil {
			return ErrInternalServer(c, err)
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
