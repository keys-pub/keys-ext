package server

import (
	"context"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/http"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// auth parameters for request.
type auth struct {
	Header string

	// Optional
	Param   string
	Content []byte

	BaseURL    string
	Now        time.Time
	NonceCheck http.NonceCheck
}

// newAuth returns new auth parameters.
func newAuth(header string, param string, content []byte) *auth {
	return &auth{
		Header:  header,
		Param:   param,
		Content: content,
	}
}

// // skipNonceCheck to skip nonce check.
// func (a *auth) skipNonceCheck() *auth {
// 	a.NonceCheck = nonceCheckSkip()
// 	return a
// }

func (s *Server) auth(c echo.Context, auth *auth) (*http.AuthResult, error) {
	request := c.Request()

	if auth.Now.IsZero() {
		auth.Now = s.clock.Now()
	}
	if auth.NonceCheck == nil {
		auth.NonceCheck = nonceCheck(s.rds)
	}
	if auth.BaseURL == "" {
		auth.BaseURL = s.URL
	}

	authReq, err := authRequest(c, auth)
	if err != nil {
		return nil, err
	}
	res, err := http.Authorize(request.Context(), authReq)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func authRequest(c echo.Context, auth *auth) (*http.AuthRequest, error) {
	request := c.Request()
	url := auth.BaseURL + c.Request().URL.String()
	contentHash := http.ContentHash(auth.Content)

	var kid keys.ID
	if auth.Param != "" {
		k, err := keys.ParseID(c.Param(auth.Param))
		if err != nil {
			return nil, errors.Wrapf(err, "invalid param")
		}
		kid = k
	}
	if auth.Header == "" {
		return nil, errors.Errorf("no auth header name, specified")
	}
	val := request.Header.Get(auth.Header)
	if val == "" {
		return nil, errors.Errorf("missing %s header", auth.Header)
	}

	return &http.AuthRequest{
		Method:      request.Method,
		URL:         url,
		KID:         kid,
		Auth:        val,
		ContentHash: contentHash,
		Now:         auth.Now,
		NonceCheck:  auth.NonceCheck,
	}, nil
}

func nonceCheck(rds Redis) http.NonceCheck {
	return func(ctx context.Context, nonce string) error {
		val, err := rds.Get(ctx, nonce)
		if err != nil {
			return err
		}
		if val != "" {
			return errors.Errorf("nonce collision")
		}
		if err := rds.Set(ctx, nonce, "1"); err != nil {
			return err
		}
		if err := rds.Expire(ctx, nonce, time.Hour); err != nil {
			return err
		}
		return nil
	}
}

// func nonceCheckSkip() http.NonceCheck {
// 	return func(ctx context.Context, nonce string) error {
// 		return nil
// 	}
// }
