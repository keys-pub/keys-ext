package server

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/keys-pub/keys"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func discoKey(kid keys.ID, rid keys.ID, typ string) string {
	addr := kid.String()
	if kid != rid {
		a, err := keys.NewAddress(kid, rid)
		if err != nil {
			panic(err)
		}
		addr = a.String()
	}
	return fmt.Sprintf("d-%s-%s", addr, shortType(typ))
}

func (s *Server) putDisco(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	kid, status, err := authorize(c, s.URL, "kid", s.nowFn(), s.mc)
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

	typ := c.Param("type")
	if typ == "" {
		return ErrBadRequest(c, errors.Errorf("no type"))
	}
	if typ != "offer" && typ != "answer" {
		return ErrBadRequest(c, errors.Errorf("invalid type"))
	}

	expire := time.Second * 15
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
	if expire > time.Minute {
		return ErrBadRequest(c, errors.Errorf("max expire is 1m"))
	}

	if c.Request().Body == nil {
		return ErrBadRequest(c, errors.Errorf("missing body"))
	}

	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return s.internalError(c, err)
	}

	if len(b) > 256 {
		// TODO: Check length before reading data
		return ErrBadRequest(c, errors.Errorf("message too large (greater than 256 bytes)"))
	}

	key := discoKey(kid, rid, typ)
	if err := s.mc.Set(ctx, key, string(b)); err != nil {
		return s.internalError(c, err)
	}

	if err := s.mc.Expire(ctx, key, expire); err != nil {
		return s.internalError(c, err)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) getDisco(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	rid, status, err := authorize(c, s.URL, "rid", s.nowFn(), s.mc)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	sender := c.Param("kid")
	if sender == "" {
		return ErrBadRequest(c, errors.Errorf("no kid specified"))
	}
	kid, err := keys.ParseID(sender)
	if err != nil {
		return ErrBadRequest(c, err)
	}

	typ := c.Param("type")
	if typ == "" {
		return ErrBadRequest(c, errors.Errorf("no type"))
	}
	if typ != "offer" && typ != "answer" {
		return ErrBadRequest(c, errors.Errorf("invalid type"))
	}

	key := discoKey(kid, rid, typ)
	out, err := s.mc.Get(ctx, key)
	if err != nil {
		return s.internalError(c, err)
	}
	if out == "" {
		return ErrNotFound(c, nil)
	}
	// Delete after get
	if err := s.mc.Delete(ctx, key); err != nil {
		return s.internalError(c, err)
	}
	return c.Blob(http.StatusOK, echo.MIMEOctetStream, []byte(out))
}

func (s *Server) deleteDisco(c echo.Context) error {
	ctx := c.Request().Context()
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	kid, status, err := authorize(c, s.URL, "kid", s.nowFn(), s.mc)
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

	okey := discoKey(kid, rid, "offer")
	if err := s.mc.Delete(ctx, okey); err != nil {
		return s.internalError(c, err)
	}
	akey := discoKey(kid, rid, "answer")
	if err := s.mc.Delete(ctx, akey); err != nil {
		return s.internalError(c, err)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

func shortType(s string) string {
	switch s {
	case "offer":
		return "o"
	case "answer":
		return "a"
	}
	panic(errors.Errorf("invalid type"))
}
