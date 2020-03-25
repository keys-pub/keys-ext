package server

import (
	"net/http"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/api"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func authorize(c echo.Context, baseURL string, now time.Time, mc MemCache) (keys.ID, int, error) {
	request := c.Request()
	auth := request.Header.Get("Authorization")
	if auth == "" {
		return "", http.StatusUnauthorized, errors.Errorf("missing Authorization header")
	}

	url := baseURL + c.Request().URL.String()

	authRes, err := api.CheckAuthorization(request.Context(), request.Method, url, auth, mc, now)
	if err != nil {
		return "", http.StatusForbidden, err
	}
	kidAuth := authRes.KID

	if c.Param("kid") != "" {
		kid, err := keys.ParseID(c.Param("kid"))
		if err != nil {
			return "", http.StatusBadRequest, err
		}

		if kid != kidAuth {
			return "", http.StatusForbidden, errors.Errorf("invalid kid")
		}
	}

	return kidAuth, 0, nil
}
