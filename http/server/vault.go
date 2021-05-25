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

	auth, err := s.auth(c, newAuthRequest("Authorization", "vid", nil))
	if err != nil {
		return s.ErrForbidden(c, err)
	}

	deleted, err := s.isVaultDeleted(c, auth.KID)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	if deleted {
		return s.ErrNotFound(c, errVaultDeleted)
	}

	limit := 1000
	path := dstore.Path("vaults", auth.KID)
	resp, err := s.events(c, path, limit)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	if len(resp.Events) == 0 && resp.Index == 0 {
		return s.ErrNotFound(c, errVaultNotFound)
	}
	truncated := false
	if len(resp.Events) >= limit {
		// TODO: This is a lie if the number of results are exactly equal to limit
		truncated = true
	}

	out := &VaultResponse{
		Vault:     resp.Events,
		Index:     resp.Index,
		Truncated: truncated,
	}

	return JSON(c, http.StatusOK, out)
}

func (s *Server) postVault(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	// TODO: max vault size

	if c.Request().Body == nil {
		return s.ErrBadRequest(c, errors.Errorf("no body data"))
	}
	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return s.ErrResponse(c, err)
	}

	auth, err := s.auth(c, newAuthRequest("Authorization", "vid", b))
	if err != nil {
		return s.ErrForbidden(c, err)
	}

	deleted, err := s.isVaultDeleted(c, auth.KID)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	if deleted {
		return s.ErrNotFound(c, errVaultDeleted)
	}

	total := int64(0)
	// JSON format uses api.Data.
	var req []*api.Data
	if err := json.Unmarshal(b, &req); err != nil {
		return s.ErrBadRequest(c, err)
	}
	docs := make([]events.Document, 0, len(req))
	for _, d := range req {
		docs = append(docs, dstore.Data(d.Data))
	}

	path := dstore.Path("vaults", auth.KID)

	if _, err := s.fi.EventsAdd(ctx, path, docs); err != nil {
		return err
	}

	// Increment usage
	for _, d := range req {
		total += int64(len(d.Data))
	}
	if _, _, err := s.fi.Increment(ctx, path, "usage", total); err != nil {
		return s.ErrResponse(c, err)
	}

	// If we have a vault token, notify.
	doc, err := s.fi.Get(ctx, path)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	if doc != nil {
		var vault Vault
		if err := doc.To(&vault); err != nil {
			return s.ErrResponse(c, err)
		}
	}
	var out struct{}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) deleteVault(c echo.Context) error {
	ctx := c.Request().Context()
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	auth, err := s.auth(c, newAuthRequest("Authorization", "vid", nil))
	if err != nil {
		return s.ErrForbidden(c, err)
	}

	deleted, err := s.isVaultDeleted(c, auth.KID)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	if deleted {
		return s.ErrNotFound(c, errVaultDeleted)
	}

	if err := s.setVaultDeleted(c, auth.KID); err != nil {
		return s.ErrResponse(c, err)
	}

	cpath := dstore.Path("vaults", auth.KID)
	exists, err := s.fi.EventsDelete(ctx, cpath)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	if !exists {
		return s.ErrNotFound(c, errVaultNotFound)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) headVault(c echo.Context) error {
	ctx := c.Request().Context()
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	auth, err := s.auth(c, newAuthRequest("Authorization", "vid", nil))
	if err != nil {
		return s.ErrForbidden(c, err)
	}

	deleted, err := s.isVaultDeleted(c, auth.KID)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	if deleted {
		return s.ErrNotFound(c, errVaultDeleted)
	}

	path := dstore.Path("vaults", auth.KID)
	iter, err := s.fi.Events(ctx, path, events.Limit(1))
	if err != nil {
		return s.ErrResponse(c, err)
	}
	defer iter.Release()
	event, err := iter.Next()
	if err != nil {
		return s.ErrResponse(c, err)
	}
	if event == nil {
		return s.ErrNotFound(c, errVaultNotFound)
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

type Vault struct {
	ID keys.ID `json:"id" msgpack:"id"`

	Index     int64 `json:"idx,omitempty" msgpack:"idx,omitempty"`
	Timestamp int64 `json:"ts,omitempty" msgpack:"ts,omitempty"`
}

// VaultResponse ...
type VaultResponse struct {
	Vault     []*api.Event `json:"vault" msgpack:"vault"`
	Index     int64        `json:"idx" msgpack:"idx"`
	Truncated bool         `json:"truncated,omitempty" msgpack:"trunc,omitempty"`
}
