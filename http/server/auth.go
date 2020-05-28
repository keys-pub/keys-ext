package server

import (
	"net/http"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func checkAuth(c echo.Context, baseURL string, now time.Time, mc MemCache) (*api.AuthResult, int, error) {
	request := c.Request()
	auth := request.Header.Get("Authorization")
	if auth == "" {
		return nil, http.StatusUnauthorized, errors.Errorf("missing Authorization header")
	}

	url := baseURL + c.Request().URL.String()

	authRes, err := api.CheckAuthorization(request.Context(), request.Method, url, auth, mc, now)
	if err != nil {
		return nil, http.StatusForbidden, err
	}
	return authRes, 0, nil
}

func authorize(c echo.Context, baseURL string, param string, now time.Time, mc MemCache) (keys.ID, int, error) {
	authRes, status, err := checkAuth(c, baseURL, now, mc)
	if err != nil {
		return "", status, err
	}
	kidAuth := authRes.KID

	if c.Param(param) != "" {
		kid, err := keys.ParseID(c.Param(param))
		if err != nil {
			return "", http.StatusBadRequest, err
		}

		if kid != kidAuth {
			return "", http.StatusForbidden, errors.Errorf("invalid " + param)
		}
	}

	return kidAuth, 0, nil
}
