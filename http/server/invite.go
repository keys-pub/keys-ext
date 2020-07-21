package server

import (
	"encoding/json"
	"fmt"
	"net/http"
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

func (s *Server) postInvite(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	kid, status, err := authorize(c, s.URL, "kid", nil, s.clock.Now(), s.rds)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}
	recipient := c.Param("rid")
	if recipient == "" {
		return ErrBadRequest(c, errors.Errorf("no recipient id"))
	}
	rid, err := keys.ParseID(recipient)
	if err != nil {
		return ErrBadRequest(c, err)
	}

	inv := invite{
		Sender:    kid,
		Recipient: rid,
	}
	ib, err := json.Marshal(inv)
	if err != nil {
		return s.internalError(c, err)
	}

	var code string
	for i := 0; i < 3; i++ {
		randWords := keys.RandWords(3)
		existing, err := s.rds.Get(ctx, code)
		if err != nil {
			return s.internalError(c, err)
		}
		if existing != "" {
			s.logger.Errorf("invite code conflict")
			continue
		}
		code = randWords
		break
	}
	if code == "" {
		return s.internalError(c, errors.Errorf("invite code conflict"))
	}

	codeKey := fmt.Sprintf("code %s", code)
	if err := s.rds.Set(ctx, codeKey, string(ib)); err != nil {
		return s.internalError(c, err)
	}
	// TODO: Configurable expiry?
	if err := s.rds.Expire(ctx, codeKey, time.Hour); err != nil {
		return s.internalError(c, err)
	}

	s.logger.Debugf("Created code: %s", code)
	resp := api.CreateInviteResponse{
		Code: code,
	}

	return JSON(c, http.StatusOK, resp)
}

func (s *Server) getInvite(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	res, status, err := checkAuth(c, s.URL, "", nil, s.clock.Now(), s.rds)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	key := fmt.Sprintf("code %s", c.QueryParam("code"))
	s.logger.Debugf("Get code: %s", key)
	out, err := s.rds.Get(ctx, key)
	if err != nil {
		return s.internalError(c, err)
	}
	if out == "" {
		return ErrNotFound(c, errors.Errorf("code not found"))
	}
	var inv invite
	if err := json.Unmarshal([]byte(out), &inv); err != nil {
		return s.internalError(c, err)
	}

	// Only allow the sender or recipient to view the invite.
	// This can happen if client has many keys and is brute forcing to find
	// which one to use.
	if inv.Recipient != res.KID && inv.Sender != res.KID {
		s.logger.Debugf("Recipient mistmatch: %s != %s", inv.Recipient, res.KID)
		return ErrNotFound(c, errors.Errorf("code not found"))
	}
	// TODO: Remove on access or when it's used?
	// if err := s.rds.Delete(ctx, key); err != nil {
	// 	return s.internalError(c, err)
	// }

	resp := api.InviteResponse{
		Sender:    inv.Sender,
		Recipient: inv.Recipient,
	}

	return JSON(c, http.StatusOK, resp)
}
