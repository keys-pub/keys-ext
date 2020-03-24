package server

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/encoding"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) putEphem(c echo.Context) error {
	request := c.Request()
	ctx := request.Context()
	logger.Infof(ctx, "Server PUT ephem %s", s.urlWithBase(c))

	kid, status, err := s.authorize(c)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	recipient := c.Param("rid")
	if recipient == "" {
		return ErrBadRequest(c, errors.Errorf("no recipient id specified"))
	}
	rid, err := keys.ParseID(recipient)
	if err != nil {
		return ErrBadRequest(c, err)
	}

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

	addr, err := keys.NewAddress(kid, rid)
	if err != nil {
		return internalError(c, err)
	}

	enc, err := encoding.Encode(bin, encoding.Base64)
	if err != nil {
		return internalError(c, err)
	}

	key := fmt.Sprintf("ephem-%s-%s", addr, id)
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

func (s *Server) getEphem(c echo.Context) error {
	request := c.Request()
	ctx := request.Context()
	logger.Infof(ctx, "Server GET ephem %s", s.urlWithBase(c))

	kid, status, err := s.authorize(c)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	recipient := c.Param("rid")
	if recipient == "" {
		return ErrBadRequest(c, errors.Errorf("no recipient id specified"))
	}
	rid, err := keys.ParseID(recipient)
	if err != nil {
		return ErrBadRequest(c, err)
	}

	id := c.Param("id")
	if id == "" {
		return ErrBadRequest(c, errors.Errorf("no id specified"))
	}

	addr, err := keys.NewAddress(kid, rid)
	if err != nil {
		return internalError(c, err)
	}

	key := fmt.Sprintf("ephem-%s-%s", addr, id)
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
