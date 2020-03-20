package server

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/keys-pub/keys/encoding"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) postRelay(c echo.Context) error {
	request := c.Request()
	ctx := request.Context()
	logger.Infof(ctx, "Server POST relay %s", s.urlString(c))
	return s.toRelay(c)
}

func (s *Server) putRelay(c echo.Context) error {
	request := c.Request()
	ctx := request.Context()
	logger.Infof(ctx, "Server PUT relay %s", s.urlString(c))
	return s.toRelay(c)
}

func (s *Server) toRelay(c echo.Context) error {
	request := c.Request()
	ctx := request.Context()

	id := c.Param("id")
	if id == "" {
		return ErrBadRequest(c, errors.Errorf("no id specified"))
	}

	if c.Request().Body == nil {
		return ErrBadRequest(c, errors.Errorf("missing body"))
	}

	bin, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return internalError(c, err)
	}

	if len(bin) > 512*1024 {
		return ErrBadRequest(c, errors.Errorf("too much data (greater than 512KiB)"))
	}

	enc, err := encoding.Encode(bin, encoding.Base64)
	if err != nil {
		return internalError(c, err)
	}

	key := fmt.Sprintf("relay-%s", id)
	if err := s.mc.Set(ctx, key, enc); err != nil {
		return internalError(c, err)
	}
	// TODO: Configurable expiry?
	if err := s.mc.Expire(ctx, key, time.Minute); err != nil {
		return internalError(c, err)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) getRelay(c echo.Context) error {
	request := c.Request()
	ctx := request.Context()
	logger.Infof(ctx, "Server GET relay %s", s.urlString(c))

	id := c.Param("id")
	if id == "" {
		return ErrBadRequest(c, errors.Errorf("no id specified"))
	}

	key := fmt.Sprintf("relay-%s", id)
	out, err := s.mc.Get(ctx, key)
	if err != nil {
		return internalError(c, err)
	}
	if out == "" {
		return ErrNotFound(c, nil)
	}
	if err := s.mc.Expire(ctx, key, time.Duration(0)); err != nil {
		return internalError(c, err)
	}

	b, err := encoding.Decode(out, encoding.Base64)
	if err != nil {
		return internalError(c, err)
	}
	return c.Blob(http.StatusOK, echo.MIMEOctetStream, b)
}
