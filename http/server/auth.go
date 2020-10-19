package server

import (
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/http"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func checkAuth(c echo.Context, baseURL string, kid keys.ID, content []byte, now time.Time, rds Redis) (*http.AuthResult, int, error) {
	request := c.Request()
	auth := request.Header.Get("Authorization")
	if auth == "" {
		return nil, http.StatusUnauthorized, errors.Errorf("missing Authorization header")
	}

	url := baseURL + c.Request().URL.String()
	contentHash := http.ContentHash(content)
	res, err := http.CheckAuthorization(request.Context(), request.Method, url, kid, auth, contentHash, rds, now)
	if err != nil {
		return nil, http.StatusForbidden, err
	}
	return res, 0, nil
}

func authorize(c echo.Context, baseURL string, param string, content []byte, now time.Time, rds Redis) (keys.ID, int, error) {
	kid, err := keys.ParseID(c.Param(param))
	if err != nil {
		return "", http.StatusBadRequest, err
	}
	res, status, err := checkAuth(c, baseURL, kid, content, now, rds)
	if err != nil {
		return "", status, err
	}
	if kid != res.KID {
		return "", http.StatusForbidden, errors.Errorf("kid mismatch")
	}
	return kid, 0, nil
}
