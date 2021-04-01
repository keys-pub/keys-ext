package server

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	wsapi "github.com/keys-pub/keys-ext/ws/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/dstore/events"
	"github.com/keys-pub/keys/tsutil"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

var errVaultNotFound = errors.New("vault not found")
var errVaultDeleted = errors.New("vault was deleted")

func (s *Server) putVault(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	body, err := readBody(c, false, 64*1024)
	if err != nil {
		return s.ErrResponse(c, err)
	}

	auth, _, err := s.auth(c, newAuthRequest("Authorization", "", body))
	if err != nil {
		return s.ErrForbidden(c, err)
	}

	vid, err := keys.ParseID(c.Param("vid"))
	if err != nil {
		return s.ErrBadRequest(c, err)
	}

	acct, err := s.findAccount(ctx, auth.KID)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	if acct == nil {
		return s.ErrForbidden(c, errors.Errorf("no account"))
	}
	// if !acct.VerifiedEmail {
	// 	return s.ErrForbidden(c, errors.Errorf("account email is not verified"))
	// }

	// Check if existing
	existing, err := s.vaultInfo(ctx, vid)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	if existing != nil && existing.Token != "" {
		return JSON(c, http.StatusOK, existing)
	}

	vaultCount, _, err := s.fi.Increment(ctx, dstore.Path(accountsCollection, auth.KID), "vaultCount", 1)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	if vaultCount > 500 {
		return s.ErrForbidden(c, errors.Errorf("max vaults reached"))
	}

	token, err := s.GenerateToken()
	if err != nil {
		return s.ErrResponse(c, err)
	}

	vault := &Vault{
		ID:    vid,
		Token: token,
	}
	path := dstore.Path("vaults", auth.KID)
	if err := s.fi.Set(ctx, path, dstore.From(vault), dstore.MergeAll()); err != nil {
		return s.ErrResponse(c, err)
	}

	// Save account vault
	av := &api.AccountVault{
		AID: auth.KID,
		VID: vid,
	}
	accountPath := dstore.Path(accountsCollection, auth.KID, "vaults", vid)
	if err := s.fi.Create(ctx, accountPath, dstore.From(av)); err != nil {
		return s.ErrResponse(c, err)
	}

	out := &VaultInfo{KID: auth.KID, Token: token}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) getVaultInfo(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	auth, _, err := s.auth(c, newAuthRequest("Authorization", "vid", nil))
	if err != nil {
		return s.ErrForbidden(c, err)
	}

	info, err := s.vaultInfo(ctx, auth.KID)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	if info == nil {
		return s.ErrNotFound(c, errors.Errorf("vault not found"))
	}
	return JSON(c, http.StatusOK, info)
}

type VaultInfo struct {
	KID   keys.ID `json:"kid"`
	Token string  `json:"token"`
	Usage int64   `json:"usage"`
}

func (s *Server) vaultInfo(ctx context.Context, kid keys.ID) (*VaultInfo, error) {
	path := dstore.Path("vaults", kid)
	var vault Vault
	ok, err := s.fi.Load(ctx, path, &vault)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	return &VaultInfo{KID: kid, Token: vault.Token, Usage: vault.Usage}, nil
}

func (s *Server) vaults(ctx context.Context, kids []keys.ID) (map[keys.ID]*Vault, error) {
	m := map[keys.ID]*Vault{}
	paths := []string{}
	for _, kid := range kids {
		paths = append(paths, dstore.Path("vaults", kid))
	}
	docs, err := s.fi.GetAll(ctx, paths)
	if err != nil {
		return nil, err
	}
	for _, doc := range docs {
		var vault Vault
		if err := doc.To(&vault); err != nil {
			return nil, err
		}
		m[vault.ID] = &vault
	}
	return m, nil
}

func (s *Server) listVault(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	auth, ext, err := s.auth(c, newAuthRequest("Authorization", "vid", nil))
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

	out := &api.VaultResponse{
		Vault:     resp.Events,
		Index:     resp.Index,
		Truncated: truncated,
	}

	switch ext {
	case "msgpack":
		return Msgpack(c, http.StatusOK, out)
	default:
		return JSON(c, http.StatusOK, out)
	}
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

	auth, ext, err := s.auth(c, newAuthRequest("Authorization", "vid", b))
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

	var data [][]byte
	total := int64(0)
	switch ext {
	case "msgpack":
		// Msgpack uses an array of bytes
		if err := msgpack.Unmarshal(b, &data); err != nil {
			return s.ErrBadRequest(c, err)
		}
	default:
		// JSON format uses api.Data.
		var req []*api.Data
		if err := json.Unmarshal(b, &req); err != nil {
			return s.ErrBadRequest(c, err)
		}
		data = make([][]byte, 0, len(req))
		for _, d := range req {
			data = append(data, d.Data)
		}
	}

	path := dstore.Path("vaults", auth.KID)

	_, idx, err := s.fi.EventsAdd(ctx, path, data)
	if err != nil {
		return err
	}

	// Increment usage
	for _, d := range data {
		total += int64(len(d))
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
		vt := &api.VaultToken{KID: auth.KID, Token: vault.Token}
		if err := s.notifyEvent(ctx, vt, idx); err != nil {
			return err
		}
	}
	var out struct{}
	return JSON(c, http.StatusOK, out)
}

func (s *Server) deleteVault(c echo.Context) error {
	ctx := c.Request().Context()
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	auth, _, err := s.auth(c, newAuthRequest("Authorization", "vid", nil))
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

	auth, _, err := s.auth(c, newAuthRequest("Authorization", "vid", nil))
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

func (s *Server) postVaultsStatus(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	body, err := readBody(c, false, 64*1024)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	var req api.VaultsStatusRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return s.ErrBadRequest(c, errors.Errorf("invalid request"))
	}
	paths := []string{}
	for k := range req.Vaults {
		kid, err := keys.ParseID(string(k))
		if err != nil {
			return s.ErrBadRequest(c, errors.Errorf("invalid request"))
		}
		paths = append(paths, dstore.Path("vaults", kid))
	}

	docs, err := s.fi.GetAll(ctx, paths)
	if err != nil {
		return s.ErrResponse(c, err)
	}
	positions, err := s.fi.EventPositions(ctx, paths)
	if err != nil {
		return s.ErrResponse(c, err)
	}

	vaults := make([]*api.VaultStatus, 0, len(docs))
	for _, doc := range docs {
		var vault Vault
		if err := doc.To(&vault); err != nil {
			return s.ErrResponse(c, err)
		}
		token := req.Vaults[vault.ID]
		if token == "" {
			s.logger.Infof("Missing token for vault %s", vault.ID)
			continue
		}
		if token != vault.Token {
			s.logger.Infof("Invalid token for vault %s", vault.ID)
			continue
		}
		vault.Timestamp = tsutil.Millis(doc.UpdatedAt)
		position := positions[doc.Path]
		if position != nil {
			vault.Index = position.Index
			if position.Timestamp > 0 {
				vault.Timestamp = position.Timestamp
			}
		}
		vaults = append(vaults, &api.VaultStatus{
			ID:        vault.ID,
			Index:     vault.Index,
			Timestamp: vault.Timestamp,
		})
	}

	out := api.VaultsStatusResponse{
		Vaults: vaults,
	}
	return c.JSON(http.StatusOK, out)
}

type Vault struct {
	ID keys.ID `json:"id" msgpack:"id"`

	Index     int64  `json:"idx,omitempty" msgpack:"idx,omitempty"`
	Timestamp int64  `json:"ts,omitempty" msgpack:"ts,omitempty"`
	Token     string `json:"token,omitempty" msgpack:"token,omitempty"`

	Usage int64 `json:"usage,omitempty" msgpack:"usage,omitempty"`
}

func (s *Server) notifyEvent(ctx context.Context, vt *api.VaultToken, idx int64) error {
	if s.internalKey == nil {
		return errors.Errorf("no secret key set")
	}
	event := &wsapi.Event{
		KID:   vt.KID,
		Index: idx,
		Token: vt.Token,
	}
	b, err := wsapi.Encrypt(event, s.internalKey)
	if err != nil {
		return err
	}
	if err := s.rds.Publish(ctx, wsapi.EventPubSub, b); err != nil {
		return err
	}
	return nil
}
