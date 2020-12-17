package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/dstore/events"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

var errVaultNotFound = errors.New("vault not found")
var errVaultDeleted = errors.New("vault was deleted")

func (s *Server) listVault(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	auth, err := s.auth(c, newAuth("Authorization", "kid", nil))
	if err != nil {
		return ErrForbidden(c, err)
	}

	deleted, err := s.isVaultDeleted(c, auth.KID)
	if err != nil {
		return ErrInternalServer(c, err)
	}
	if deleted {
		return ErrNotFound(c, errVaultDeleted)
	}

	limit := 1000
	path := dstore.Path("vaults", auth.KID)
	resp, st, err := s.events(c, path, limit)
	if err != nil {
		return ErrResponse(c, st, err)
	}
	if len(resp.Events) == 0 && resp.Index == 0 {
		return ErrNotFound(c, errVaultNotFound)
	}
	truncated := false
	if len(resp.Events) >= limit {
		// TODO: This is a lie if the number of results are exactly equal to limit
		truncated = true
	}

	out := &api.VaultResponse{
		Vault:     resp.Events,
		Index:     resp.Index,
		Truncated: truncated,
	}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) postVault(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	// TODO: max vault size

	if c.Request().Body == nil {
		return ErrBadRequest(c, errors.Errorf("no body data"))
	}
	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return ErrInternalServer(c, err)
	}

	auth, err := s.auth(c, newAuth("Authorization", "kid", b))
	if err != nil {
		return ErrForbidden(c, err)
	}

	deleted, err := s.isVaultDeleted(c, auth.KID)
	if err != nil {
		return ErrInternalServer(c, err)
	}
	if deleted {
		return ErrNotFound(c, errVaultDeleted)
	}

	var req []*api.Data
	if err := json.Unmarshal(b, &req); err != nil {
		return ErrBadRequest(c, err)
	}

	ctx := c.Request().Context()
	cpath := dstore.Path("vaults", auth.KID)
	data := make([][]byte, 0, len(req))
	for _, d := range req {
		data = append(data, d.Data)
	}
	if _, _, err := s.fi.EventsAdd(ctx, cpath, data); err != nil {
		return ErrInternalServer(c, err)
	}

	var out struct{}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) deleteVault(c echo.Context) error {
	ctx := c.Request().Context()
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	auth, err := s.auth(c, newAuth("Authorization", "kid", nil))
	if err != nil {
		return ErrForbidden(c, err)
	}

	deleted, err := s.isVaultDeleted(c, auth.KID)
	if err != nil {
		return ErrInternalServer(c, err)
	}
	if deleted {
		return ErrNotFound(c, errVaultDeleted)
	}

	if err := s.setVaultDeleted(c, auth.KID); err != nil {
		return ErrInternalServer(c, err)
	}

	cpath := dstore.Path("vaults", auth.KID)
	exists, err := s.fi.EventsDelete(ctx, cpath)
	if err != nil {
		return ErrInternalServer(c, err)
	}
	if !exists {
		return ErrNotFound(c, errVaultNotFound)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) headVault(c echo.Context) error {
	ctx := c.Request().Context()
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	auth, err := s.auth(c, newAuth("Authorization", "kid", nil))
	if err != nil {
		return ErrForbidden(c, err)
	}

	deleted, err := s.isVaultDeleted(c, auth.KID)
	if err != nil {
		return ErrInternalServer(c, err)
	}
	if deleted {
		return ErrNotFound(c, errVaultDeleted)
	}

	path := dstore.Path("vaults", auth.KID)
	iter, err := s.fi.Events(ctx, path, events.Limit(1))
	if err != nil {
		return ErrInternalServer(c, err)
	}
	defer iter.Release()
	event, err := iter.Next()
	if err != nil {
		return ErrInternalServer(c, err)
	}
	if event == nil {
		return ErrNotFound(c, errVaultNotFound)
	}

	return c.NoContent(http.StatusOK)
}

func (s Server) isVaultDeleted(c echo.Context, kid keys.ID) (bool, error) {
	ctx := c.Request().Context()
	return s.fi.Exists(ctx, dstore.Path("vaults-rm", kid))
}

func (s *Server) setVaultDeleted(c echo.Context, kid keys.ID) error {
	ctx := c.Request().Context()
	return s.fi.Set(ctx, dstore.Path("vaults-rm", kid), dstore.Data([]byte{}))
}
