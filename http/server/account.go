package server

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"time"

	"github.com/badoux/checkmail"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

var accountsCollection = "accounts"

func (s *Server) putAccount(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	body, err := readBody(c, false, 64*1024)
	if err != nil {
		return s.ErrResponse(c, err)
	}

	auth, _, err := s.auth(c, newAuthRequest("Authorization", "aid", body))
	if err != nil {
		return s.ErrForbidden(c, err)
	}

	var req api.AccountCreateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return s.ErrBadRequest(c, err)
	}
	if err := checkmail.ValidateFormat(req.Email); err != nil {
		return s.ErrBadRequest(c, errors.Errorf("invalid email"))
	}

	existing, err := s.findAccountByEmail(ctx, req.Email)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	if existing != nil {
		return s.ErrConflict(c, errors.Errorf("account already exists"))
	}

	path := dstore.Path(accountsCollection, auth.KID)

	acct := &api.Account{
		Email: req.Email,
		KID:   auth.KID,
	}

	if err := s.fi.Create(ctx, path, dstore.From(acct)); err != nil {
		switch err.(type) {
		case dstore.ErrPathExists:
			return s.ErrConflict(c, errors.Errorf("account already exists"))
		}
		return s.ErrResponse(c, err)
	}

	if err := s.sendEmailVerification(c, acct); err != nil {
		return s.ErrResponse(c, err)
	}

	out := &api.AccountCreateResponse{
		Email: acct.Email,
		KID:   acct.KID,
	}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) postAccountSendVerifyEmail(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	auth, _, err := s.auth(c, newAuthRequest("Authorization", "aid", nil))
	if err != nil {
		return s.ErrForbidden(c, err)
	}

	acct, err := s.findAccount(ctx, auth.KID)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	if acct == nil {
		return s.ErrNotFound(c, keys.NewErrNotFound(auth.KID.String()))
	}
	if acct.VerifiedEmail {
		return s.ErrBadRequest(c, errors.Errorf("already verified"))
	}
	if s.clock.Now().Sub(acct.VerifyEmailCodeAt) > time.Minute {
		return s.ErrTooManyRequests(c, errors.Errorf("already sent verification recently"))
	}

	if err := s.sendEmailVerification(c, acct); err != nil {
		return s.ErrResponse(c, err)
	}

	out := &api.SendEmailVerificationResponse{
		Email: acct.Email,
		KID:   acct.KID,
	}
	return c.JSON(http.StatusOK, out)
}

func (s *Server) getAccount(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	auth, _, err := s.auth(c, newAuthRequest("Authorization", "aid", nil))
	if err != nil {
		return s.ErrForbidden(c, err)
	}

	acct, err := s.findAccount(ctx, auth.KID)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	if acct == nil {
		return s.ErrNotFound(c, keys.NewErrNotFound(auth.KID.String()))
	}

	out := &api.AccountResponse{
		Email:         acct.Email,
		KID:           acct.KID,
		VerifiedEmail: acct.VerifiedEmail,
	}
	return c.JSON(http.StatusOK, out)
}

func (s *Server) findAccount(ctx context.Context, kid keys.ID) (*api.Account, error) {
	path := dstore.Path(accountsCollection, kid)

	doc, err := s.fi.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}

	var acct api.Account
	if err := doc.To(&acct); err != nil {
		return nil, err
	}

	return &acct, nil
}

func (s *Server) findAccountByEmail(ctx context.Context, email string) (*api.Account, error) {
	docs, err := s.fi.Documents(ctx, dstore.Path(accountsCollection), dstore.Where("email", "==", email))
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return nil, nil
	}
	var acct api.Account
	if err := docs[0].To(&acct); err != nil {
		return nil, err
	}
	return &acct, nil
}

func (s *Server) sendEmailVerification(c echo.Context, acct *api.Account) error {
	ctx := c.Request().Context()

	verifyCode := keys.RandDigits(6)
	update := struct {
		VerifyEmailCode   string    `json:"verifyEmailCode"`
		VerifyEmailCodeAt time.Time `json:"verifyEmailCodeAt"`
	}{
		VerifyEmailCode:   verifyCode,
		VerifyEmailCodeAt: s.clock.Now(),
	}

	path := dstore.Path(accountsCollection, acct.KID)
	if err := s.fi.Set(ctx, path, dstore.From(update), dstore.MergeAll()); err != nil {
		return err
	}

	if s.emailer == nil {
		return errors.Errorf("no emailer set")
	}
	if err := s.emailer.SendVerificationEmail(acct.Email, verifyCode); err != nil {
		return err
	}
	return nil
}

func (s *Server) postAccountVerifyEmail(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	body, err := readBody(c, false, 64*1024)
	if err != nil {
		return s.ErrResponse(c, err)
	}

	auth, _, err := s.auth(c, newAuthRequest("Authorization", "aid", body))
	if err != nil {
		return s.ErrForbidden(c, err)
	}

	var req api.AccountVerifyEmailRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return s.ErrBadRequest(c, err)
	}

	acct, err := s.findAccount(ctx, auth.KID)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	if acct == nil {
		return s.ErrNotFound(c, keys.NewErrNotFound(auth.KID.String()))
	}
	if s.clock.Now().Sub(acct.VerifyEmailCodeAt) > time.Hour {
		return s.ErrBadRequest(c, errors.Errorf("expired code"))
	}
	if subtle.ConstantTimeCompare([]byte(acct.VerifyEmailCode), []byte(req.Code)) != 1 {
		return s.ErrBadRequest(c, errors.Errorf("invalid code"))
	}

	update := struct {
		Verified   bool      `json:"verifiedEmail"`
		VerifiedAt time.Time `json:"verifiedEmailAt"`
	}{
		Verified:   true,
		VerifiedAt: s.clock.Now(),
	}

	path := dstore.Path(accountsCollection, acct.KID)
	if err := s.fi.Set(ctx, path, dstore.From(update), dstore.MergeAll()); err != nil {
		return err
	}

	after, err := s.findAccount(ctx, auth.KID)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	if acct == nil {
		return s.ErrNotFound(c, keys.NewErrNotFound(auth.KID.String()))
	}

	out := &api.AccountResponse{
		Email:         after.Email,
		KID:           after.KID,
		VerifiedEmail: after.VerifiedEmail,
	}
	return c.JSON(http.StatusOK, out)
}
