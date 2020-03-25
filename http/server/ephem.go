package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keysd/http/api"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type invite struct {
	Sender    keys.ID `json:"s"`
	Recipient keys.ID `json:"r"`
}

func (s *Server) putEphem(c echo.Context) error {
	ctx := c.Request().Context()
	s.logger.Infof("Server PUT ephem %s", c.Request().URL.String())

	kid, status, err := authorize(c, s.URL, s.nowFn(), s.mc)
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

	genCode := truthy(c.QueryParam("code"))

	if c.Request().Body == nil {
		return ErrBadRequest(c, errors.Errorf("missing body"))
	}

	expire := time.Second * 15
	pexpire := c.QueryParam("expire")
	if pexpire != "" {
		expire, err = time.ParseDuration(pexpire)
		if err != nil {
			return ErrBadRequest(c, err)
		}
	}
	// TODO: Max expire

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

	key := fmt.Sprintf("ephem-%s-%s", kid, rid)
	if err := s.mc.Set(ctx, key, enc); err != nil {
		return internalError(c, err)
	}
	// TODO: Configurable expiry?
	if err := s.mc.Expire(ctx, key, expire); err != nil {
		return internalError(c, err)
	}

	code := ""
	if genCode {
		code = keys.RandWords(3)
		inv := invite{
			Sender:    kid,
			Recipient: rid,
		}
		ib, err := json.Marshal(inv)
		if err != nil {
			return internalError(c, err)
		}

		codeKey := fmt.Sprintf("code %s", code)
		if err := s.mc.Set(ctx, codeKey, string(ib)); err != nil {
			return internalError(c, err)
		}
		// TODO: Configurable expiry?
		if err := s.mc.Expire(ctx, codeKey, time.Hour); err != nil {
			return internalError(c, err)
		}
	}

	resp := api.EphemResponse{
		Code: code,
	}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) deleteEphem(c echo.Context) error {
	ctx := c.Request().Context()
	s.logger.Infof("Server PUT ephem %s", c.Request().URL.String())

	kid, status, err := authorize(c, s.URL, s.nowFn(), s.mc)
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

	key := fmt.Sprintf("ephem-%s-%s", kid, rid)
	if err := s.mc.Delete(ctx, key); err != nil {
		return internalError(c, err)
	}

	// TODO: Delete associated code too?

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) getEphem(c echo.Context) error {
	s.logger.Infof("Server GET ephem %s", c.Request().URL.String())
	ctx := c.Request().Context()

	kid, status, err := authorize(c, s.URL, s.nowFn(), s.mc)
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

	key := fmt.Sprintf("ephem-%s-%s", rid, kid)
	out, err := s.mc.Get(ctx, key)
	if err != nil {
		return internalError(c, err)
	}
	if out == "" {
		return ErrNotFound(c, nil)
	}
	if err := s.mc.Delete(ctx, key); err != nil {
		return internalError(c, err)
	}

	b, err := encoding.Decode(out, encoding.Base64)
	if err != nil {
		return internalError(c, err)
	}
	return c.Blob(http.StatusOK, echo.MIMEOctetStream, b)
}

func (s *Server) getInvite(c echo.Context) error {
	s.logger.Infof("Server GET invite %s", c.Request().URL.String())
	ctx := c.Request().Context()

	s.logger.Debugf("Auth")
	kid, status, err := authorize(c, s.URL, s.nowFn(), s.mc)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	key := fmt.Sprintf("code %s", c.QueryParam("code"))
	s.logger.Debugf("Get code: %s", key)
	out, err := s.mc.Get(ctx, key)
	if err != nil {
		return internalError(c, err)
	}
	if out == "" {
		return ErrNotFound(c, errors.Errorf("code not found"))
	}
	var inv invite
	if err := json.Unmarshal([]byte(out), &inv); err != nil {
		return internalError(c, err)
	}
	// This can happen if client has many keys and is brute forcing to find
	// which one to use.
	if inv.Recipient != kid {
		// s.logger.Errorf("Recipient mistmatch: %s != %s", inv.Recipient, kid)
		return ErrNotFound(c, errors.Errorf("code not found"))
	}
	// TODO: Remove on access or when it's used?
	// if err := s.mc.Delete(ctx, key); err != nil {
	// 	return internalError(c, err)
	// }

	resp := api.InviteResponse{
		Sender:    inv.Sender,
		Recipient: inv.Recipient,
	}

	return JSON(c, http.StatusOK, resp)
}

func truthy(s string) bool {
	s = strings.TrimSpace(s)
	switch s {
	case "", "0", "f", "false", "n", "no":
		return false
	default:
		return true
	}
}
