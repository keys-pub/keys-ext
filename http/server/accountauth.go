package server

import (
	"encoding/json"
	"io/ioutil"

	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) postAccountAuth(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	request := c.Request()
	ctx := request.Context()

	if c.Request().Body == nil {
		return s.ErrBadRequest(c, errors.Errorf("no body data"))
	}
	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return s.ErrResponse(c, err)
	}

	auth, _, err := s.auth(c, newAuthRequest("Authorization", "aid", b))
	if err != nil {
		return s.ErrForbidden(c, err)
	}

	var accountAuth api.AccountAuth
	if err := json.Unmarshal(b, &accountAuth); err != nil {
		return s.ErrBadRequest(c, err)
	}

	path := dstore.Path(accountsCollection, auth.KID, "auths", accountAuth.ID)
	if err := s.fi.Create(ctx, path, dstore.From(accountAuth)); err != nil {
		return s.ErrResponse(c, err)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) getAccountAuths(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	request := c.Request()
	ctx := request.Context()

	auth, _, err := s.auth(c, newAuthRequest("Authorization", "aid", nil))
	if err != nil {
		return s.ErrForbidden(c, err)
	}

	path := dstore.Path(accountsCollection, auth.KID, "auths")
	iter, err := s.fi.DocumentIterator(ctx, path)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	defer iter.Release()

	accountAuths := []*api.AccountAuth{}
	for {
		doc, err := iter.Next()
		if err != nil {
			return s.ErrResponse(c, err)
		}
		if doc == nil {
			break
		}
		var accountAuth api.AccountAuth
		if err := doc.To(&accountAuth); err != nil {
			return s.ErrResponse(c, err)
		}
		accountAuths = append(accountAuths, &accountAuth)
	}

	out := api.AccountAuthsResponse{Auths: accountAuths}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) deleteAuth(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	request := c.Request()
	ctx := request.Context()

	auth, _, err := s.auth(c, newAuthRequest("Authorization", "aid", nil))
	if err != nil {
		return s.ErrForbidden(c, err)
	}

	id := c.Param("id")
	if id == "" {
		return s.ErrBadRequest(c, errors.Errorf("empty id"))
	}

	path := dstore.Path(accountsCollection, auth.KID, "auths", id)
	ok, err := s.fi.Delete(ctx, path)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	if !ok {
		return s.ErrNotFound(c, errors.Errorf("auth not found"))
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}
