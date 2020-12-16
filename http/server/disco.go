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

	if c.Request().Body == nil {
		return ErrBadRequest(c, errors.Errorf("missing body"))
	}

	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return ErrInternalServer(c, err)
	}

	if len(b) > 256 {
		// TODO: Check length before reading data
		return ErrBadRequest(c, errors.Errorf("message too large (greater than 256 bytes)"))
	}

	auth, err := s.auth(c, newAuth("Authorization", "kid", b))
	if err != nil {
		return ErrForbidden(c, err)
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

	expire, err := queryParamDuration(c, "expire", time.Second*15)
	if err != nil {
		return ErrBadRequest(c, err)
	}

	if len(expire.String()) > 64 {
		return ErrBadRequest(c, errors.Errorf("invalid expire"))
	}
	if expire > time.Minute {
		return ErrBadRequest(c, errors.Errorf("max expire is 1m"))
	}

	key := discoKey(auth.KID, rid, typ)
	if err := s.rds.Set(ctx, key, string(b)); err != nil {
		return ErrInternalServer(c, err)
	}

	if err := s.rds.Expire(ctx, key, expire); err != nil {
		return ErrInternalServer(c, err)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) getDisco(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	auth, err := s.auth(c, newAuth("Authorization", "rid", nil))
	if err != nil {
		return ErrForbidden(c, err)
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

	key := discoKey(kid, auth.KID, typ)
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

func (s *Server) deleteDisco(c echo.Context) error {
	ctx := c.Request().Context()
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	auth, err := s.auth(c, newAuth("Authorization", "kid", nil))
	if err != nil {
		return ErrForbidden(c, err)
	}

	recipient := c.Param("rid")
	if recipient == "" {
		return ErrBadRequest(c, errors.Errorf("no recipient id specified"))
	}
	rid, err := keys.ParseID(recipient)
	if err != nil {
		return ErrBadRequest(c, err)
	}

	okey := discoKey(auth.KID, rid, "offer")
	if err := s.rds.Delete(ctx, okey); err != nil {
		return ErrInternalServer(c, err)
	}
	akey := discoKey(auth.KID, rid, "answer")
	if err := s.rds.Delete(ctx, akey); err != nil {
		return ErrInternalServer(c, err)
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
