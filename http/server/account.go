package server

import (
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// Account document.
type Account struct {
	UID   string  `json:"uid"`
	Token string  `json:"token"`
	KID   keys.ID `json:"kid"`
}

func (s *Server) putAccount(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	if s.firebaseAuth == nil {
		return s.ErrNotFound(c, nil)
	}

	ctx := c.Request().Context()
	email := c.FormValue("email")
	kid, err := keys.ParseID(c.FormValue("kid"))
	if err != nil {
		return s.ErrBadRequest(c, err)
	}

	token := keys.RandPassword(16)
	uid, err := s.firebaseAuth.CreateEmailUser(ctx, email, token)
	if err != nil {
		return s.ErrResponse(c, err)
	}

	path := dstore.Path("accounts", kid)

	acct := &Account{
		UID:   uid,
		Token: token,
		KID:   kid,
	}

	if err := s.fi.Create(ctx, path, dstore.From(acct)); err != nil {
		switch err.(type) {
		case dstore.ErrPathExists:
			return s.ErrConflict(c, errors.Errorf("account already exists"))
		}
		return s.ErrResponse(c, err)
	}

	var out struct{}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) getAccount(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	if s.firebaseAuth == nil {
		return s.ErrNotFound(c, nil)
	}
	// ctx := c.Request().Context()
	// header := c.Request().Header.Get("Authorization")

	// uid, err := s.firebaseAuth.VerifyIDToken(ctx, header)

	// path := dstore.Path("accounts", uid)

	// doc, err := s.fi.Get(ctx, path)
	// if err != nil {
	// 	return s.ErrResponse(c, err)
	// }
	// if doc == nil {
	// 	return s.ErrNotFound(c, nil)
	// }

	// var out Account
	// if err := doc.To(&out); err != nil {
	// 	return s.ErrResponse(c, err)
	// }
	// // out.Timestamp = tsutil.Millis(doc.UpdatedAt)

	var out struct{}
	return c.JSON(http.StatusOK, out)
}

func (s *Server) sendEmailVerification(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()
	if s.firebaseAuth == nil {
		return s.ErrNotFound(c, nil)
	}

	if err := s.firebaseAuth.SendEmailVerification(ctx, "", ""); err != nil {
		return s.ErrResponse(c, err)
	}

	var out struct{}
	return c.JSON(http.StatusOK, out)
}
