package server

import (
	"net/http"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/api"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) authorize(c echo.Context) (keys.ID, int, error) {
	request := c.Request()
	auth := request.Header.Get("Authorization")
	if auth == "" {
		return "", http.StatusUnauthorized, errors.Errorf("missing Authorization header")
	}
	now := s.nowFn()
	authRes, err := api.CheckAuthorization(request.Context(), request.Method, s.urlWithBase(c), auth, s.mc, now)
	if err != nil {
		return "", http.StatusForbidden, err
	}
	kidAuth := authRes.KID

	kid, err := keys.ParseID(c.Param("kid"))
	if err != nil {
		return "", http.StatusBadRequest, err
	}

	if kid != kidAuth {
		return "", http.StatusForbidden, errors.Errorf("invalid kid")
	}

	return kid, 0, nil
}
