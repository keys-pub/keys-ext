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

	kid, status, err := authorize(c, s.URL, "kid", s.nowFn(), s.rds)
	if err != nil {
		return ErrResponse(c, status, err.Error())
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

	if c.Request().Body == nil {
		return ErrBadRequest(c, errors.Errorf("missing body"))
	}

	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return s.internalError(c, err)
	}

	if len(b) > 512 {
		// TODO: Check length before reading data
		return ErrBadRequest(c, errors.Errorf("message too large (greater than 512 bytes)"))
	}

	key := fmt.Sprintf("s-%s", kid)
	if err := s.rds.Set(ctx, key, string(b)); err != nil {
		return s.internalError(c, err)
	}

	if err := s.rds.Expire(ctx, key, expire); err != nil {
		return s.internalError(c, err)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) getShare(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	kid, status, err := authorize(c, s.URL, "kid", s.nowFn(), s.rds)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	key := fmt.Sprintf("s-%s", kid)
	out, err := s.rds.Get(ctx, key)
	if err != nil {
		return s.internalError(c, err)
	}
	if out == "" {
		return ErrNotFound(c, nil)
	}
	// Delete after get
	if err := s.rds.Delete(ctx, key); err != nil {
		return s.internalError(c, err)
	}
	return c.Blob(http.StatusOK, echo.MIMEOctetStream, []byte(out))
}
