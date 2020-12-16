package server

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) putShare(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	if c.Request().Body == nil {
		return ErrBadRequest(c, errors.Errorf("missing body"))
	}

	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return ErrInternalServer(c, err)
	}

	if len(b) > 512 {
		// TODO: Check length before reading data
		return ErrBadRequest(c, errors.Errorf("message too large (greater than 512 bytes)"))
	}

	auth, err := s.auth(c, newAuth("Authorization", "kid", b))
	if err != nil {
		return ErrForbidden(c, err)
	}

	expire := time.Minute * 5
	if c.QueryParam("expire") != "" {
		e, err := time.ParseDuration(c.QueryParam("expire"))
		if err != nil {
			return ErrBadRequest(c, err)
		}
		expire = e
	}
	if len(expire.String()) > 64 {
		return ErrBadRequest(c, errors.Errorf("invalid expire"))
	}
	if expire > 15*time.Minute {
		return ErrBadRequest(c, errors.Errorf("max expire is 15m"))
	}

	key := fmt.Sprintf("s-%s", auth.KID)
	if err := s.rds.Set(ctx, key, string(b)); err != nil {
		return ErrInternalServer(c, err)
	}

	if err := s.rds.Expire(ctx, key, expire); err != nil {
		return ErrInternalServer(c, err)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) getShare(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	auth, err := s.auth(c, newAuth("Authorization", "kid", nil))
	if err != nil {
		return ErrForbidden(c, err)
	}

	key := fmt.Sprintf("s-%s", auth.KID)
	out, err := s.rds.Get(ctx, key)
	if err != nil {
		return ErrInternalServer(c, err)
	}
	if out == "" {
		return ErrNotFound(c, nil)
	}
	// Delete after get
	if err := s.rds.Delete(ctx, key); err != nil {
		return ErrInternalServer(c, err)
	}
	return c.Blob(http.StatusOK, echo.MIMEOctetStream, []byte(out))
}
