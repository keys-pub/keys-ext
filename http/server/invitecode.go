package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type invite struct {
	Sender    keys.ID `json:"s"`
	Recipient keys.ID `json:"r"`
}

func (s *Server) postInviteCode(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	auth, err := s.auth(c, newAuthRequest("Authorization", "kid", nil))
	if err != nil {
		return s.ErrForbidden(c, err)
	}

	recipient := c.Param("rid")
	if recipient == "" {
		return s.ErrBadRequest(c, errors.Errorf("no recipient id"))
	}
	rid, err := keys.ParseID(recipient)
	if err != nil {
		return s.ErrBadRequest(c, err)
	}

	inv := invite{
		Sender:    auth.KID,
		Recipient: rid,
	}
	ib, err := json.Marshal(inv)
	if err != nil {
		return s.ErrResponse(c, err)
	}

	var code string
	for i := 0; i < 3; i++ {
		randWords := keys.RandWords(3)
		existing, err := s.rds.Get(ctx, code)
		if err != nil {
			return s.ErrResponse(c, err)
		}
		if existing != "" {
			s.logger.Errorf("invite code conflict")
			continue
		}
		code = randWords
		break
	}
	if code == "" {
		return s.ErrResponse(c, errors.Errorf("invite code conflict"))
	}

	codeKey := fmt.Sprintf("code %s", code)
	if err := s.rds.Set(ctx, codeKey, string(ib)); err != nil {
		return s.ErrResponse(c, err)
	}
	// TODO: Configurable expiry?
	if err := s.rds.Expire(ctx, codeKey, time.Hour); err != nil {
		return s.ErrResponse(c, err)
	}

	s.logger.Debugf("Created code: %s", code)
	resp := api.InviteCodeCreateResponse{
		Code: code,
	}

	return JSON(c, http.StatusOK, resp)
}

func (s *Server) getInviteCode(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	auth, err := s.auth(c, newAuthRequest("Authorization", "", nil))
	if err != nil {
		return s.ErrForbidden(c, err)
	}

	code, err := url.QueryUnescape(c.Param("code"))
	if err != nil {
		return s.ErrBadRequest(c, err)
	}
	key := fmt.Sprintf("code %s", code)
	s.logger.Debugf("Get code: %s", key)
	out, err := s.rds.Get(ctx, key)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	if out == "" {
		return s.ErrNotFound(c, errors.Errorf("code not found"))
	}
	var inv invite
	if err := json.Unmarshal([]byte(out), &inv); err != nil {
		return s.ErrResponse(c, err)
	}

	// Only allow the sender or recipient to view the invite.
	// This can happen if client has many keys and is brute forcing to find
	// which one to use.
	if inv.Recipient != auth.KID && inv.Sender != auth.KID {
		s.logger.Debugf("Recipient mistmatch: %s != %s", inv.Recipient, auth.KID)
		return s.ErrNotFound(c, errors.Errorf("code not found"))
	}
	// TODO: Remove on access or when it's used?
	// if err := s.rds.Delete(ctx, key); err != nil {
	// 	return s.ErrResponse(c, err)
	// }

	resp := api.InviteCodeResponse{
		Sender:    inv.Sender,
		Recipient: inv.Recipient,
	}

	return JSON(c, http.StatusOK, resp)
}
