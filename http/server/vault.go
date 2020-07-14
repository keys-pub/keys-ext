package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/docs"
	"github.com/keys-pub/keys/docs/events"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

var errVaultNotFound = errors.New("vault not found")
var errVaultDeleted = errors.New("vault was deleted")

func (s *Server) postVault(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	kid, status, err := authorize(c, s.URL, "kid", s.nowFn(), s.rds)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	deleted, err := s.isVaultDeleted(c, kid)
	if err != nil {
		return s.internalError(c, err)
	}
	if deleted {
		return ErrNotFound(c, errVaultDeleted)
	}

	// TODO: max vault size

	var data []byte
	if c.Request().Body != nil {
		b, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return s.internalError(c, err)
		}
		data = b
	}

	ctx := c.Request().Context()
	cpath := docs.Path("vaults", kid)
	if _, err := s.fi.EventsAdd(ctx, cpath, [][]byte{data}); err != nil {
		return s.internalError(c, err)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) listVault(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	kid, status, err := authorize(c, s.URL, "kid", s.nowFn(), s.rds)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	deleted, err := s.isVaultDeleted(c, kid)
	if err != nil {
		return s.internalError(c, err)
	}
	if deleted {
		return ErrNotFound(c, errVaultDeleted)
	}

	cpath := docs.Path("vaults", kid)
	resp, err := s.events(c, cpath)
	if err != nil {
		return err
	}
	if len(resp.Events) == 0 && resp.Index == 0 {
		return ErrNotFound(c, errVaultNotFound)
	}

	return JSON(c, http.StatusOK, resp)
}

func (s *Server) putVault(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	kid, status, err := authorize(c, s.URL, "kid", s.nowFn(), s.rds)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	deleted, err := s.isVaultDeleted(c, kid)
	if err != nil {
		return s.internalError(c, err)
	}
	if deleted {
		return ErrNotFound(c, errVaultDeleted)
	}

	// TODO: max vault size

	if c.Request().Body == nil {
		return ErrBadRequest(c, errors.Errorf("no body data"))
	}
	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return s.internalError(c, err)
	}

	var req []*api.Data
	if err := json.Unmarshal(b, &req); err != nil {
		return ErrBadRequest(c, err)
	}

	ctx := c.Request().Context()
	cpath := docs.Path("vaults", kid)
	data := make([][]byte, 0, len(req))
	for _, d := range req {
		data = append(data, d.Data)
	}
	if _, err := s.fi.EventsAdd(ctx, cpath, data); err != nil {
		return s.internalError(c, err)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) deleteVault(c echo.Context) error {
	ctx := c.Request().Context()
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	kid, status, err := authorize(c, s.URL, "kid", s.nowFn(), s.rds)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	deleted, err := s.isVaultDeleted(c, kid)
	if err != nil {
		return s.internalError(c, err)
	}
	if deleted {
		return ErrNotFound(c, errVaultDeleted)
	}

	if err := s.setVaultDeleted(c, kid); err != nil {
		return s.internalError(c, err)
	}

	cpath := docs.Path("vaults", kid)
	exists, err := s.fi.EventsDelete(ctx, cpath)
	if err != nil {
		return s.internalError(c, err)
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

	kid, status, err := authorize(c, s.URL, "kid", s.nowFn(), s.rds)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	deleted, err := s.isVaultDeleted(c, kid)
	if err != nil {
		return s.internalError(c, err)
	}
	if deleted {
		return ErrNotFound(c, errVaultDeleted)
	}

	path := docs.Path("vaults", kid)
	iter, err := s.fi.Events(ctx, path, events.Limit(1))
	if err != nil {
		return s.internalError(c, err)
	}
	defer iter.Release()
	event, err := iter.Next()
	if err != nil {
		return s.internalError(c, err)
	}
	if event == nil {
		return ErrNotFound(c, errVaultNotFound)
	}

	return c.NoContent(http.StatusOK)
}

func (s Server) isVaultDeleted(c echo.Context, kid keys.ID) (bool, error) {
	ctx := c.Request().Context()
	return s.fi.Exists(ctx, docs.Path("vaults-del", kid))
}

func (s *Server) setVaultDeleted(c echo.Context, kid keys.ID) error {
	ctx := c.Request().Context()
	return s.fi.Set(ctx, docs.Path("vaults-del", kid), []byte("1"))
}
