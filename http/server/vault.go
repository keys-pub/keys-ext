package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/api"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

const vaultlog = "vaultlog"

func (s *Server) putVault(c echo.Context) error {
	request := c.Request()
	ctx := request.Context()
	logger.Infof(ctx, "Server PUT vault %s", s.urlString(c))

	// Auth
	auth := request.Header.Get("Authorization")
	if auth == "" {
		return ErrUnauthorized(c, errors.Errorf("missing Authorization header"))
	}
	now := s.nowFn()
	authRes, err := CheckAuthorization(request.Context(), request.Method, s.urlString(c), auth, s.mc, now)
	if err != nil {
		return ErrForbidden(c, err)
	}
	kidAuth := authRes.kid

	kid, err := keys.ParseID(c.Param("kid"))
	if err != nil {
		return ErrBadRequest(c, err)
	}

	if kid != kidAuth {
		return ErrForbidden(c, errors.Errorf("invalid kid"))
	}
	// End Auth

	logname := vaultlog + "-" + kid.String()

	id, err := keys.ParseID(c.Param("id"))
	if err != nil {
		return ErrBadRequest(c, err)
	}

	if c.Request().Body == nil {
		return ErrBadRequest(c, errors.Errorf("missing body"))
	}

	bin, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return internalError(c, err)
	}

	item := api.Item{
		ID:   id,
		Data: bin,
		Path: keys.Path("vault", id),
	}
	logger.Infof(ctx, "Set %s", item.Path)
	mb, err := json.Marshal(item)
	if err != nil {
		return internalError(c, err)
	}
	if err := s.fi.Create(ctx, item.Path, mb); err != nil {
		return internalError(c, err)
	}
	logger.Infof(ctx, "Add change %s %s", logname, item.Path)
	if err := s.fi.ChangeAdd(ctx, logname, item.Path); err != nil {
		return internalError(c, err)
	}

	return c.String(http.StatusOK, "{}")
}

func (s *Server) listVault(c echo.Context) error {
	request := c.Request()
	ctx := request.Context()
	logger.Infof(ctx, "Server PUT vault %s", s.urlString(c))

	// Auth
	auth := request.Header.Get("Authorization")
	if auth == "" {
		return ErrUnauthorized(c, errors.Errorf("missing Authorization header"))
	}
	now := s.nowFn()
	authRes, err := CheckAuthorization(request.Context(), request.Method, s.urlString(c), auth, s.mc, now)
	if err != nil {
		return ErrForbidden(c, err)
	}
	kidAuth := authRes.kid

	kid, err := keys.ParseID(c.Param("kid"))
	if err != nil {
		return ErrBadRequest(c, err)
	}

	if kid != kidAuth {
		return ErrForbidden(c, errors.Errorf("invalid kid"))
	}
	// End Auth

	logname := vaultlog + "-" + kid.String()

	le, err := s.changes(c, logname)
	if err != nil {
		return internalError(c, err)
	}
	if le.badRequest != nil {
		return le.badRequest
	}
	if len(le.docs) == 0 && le.version == 0 {
		return ErrNotFound(c, errors.Errorf("vault not found"))
	}

	items := make([]*api.Item, 0, len(le.docs))
	md := make(map[string]api.Metadata, len(le.docs))
	for _, doc := range le.docs {
		var item api.Item
		if err := json.Unmarshal(doc.Data, &item); err != nil {
			return internalError(c, err)
		}
		items = append(items, &item)
		md[item.Path] = api.Metadata{
			CreatedAt: doc.CreatedAt,
			UpdatedAt: doc.UpdatedAt,
		}
	}

	resp := api.VaultResponse{
		Items:   items,
		KID:     kid,
		Version: fmt.Sprintf("%d", le.versionNext),
	}
	fields := keys.NewStringSetSplit(c.QueryParam("include"), ",")
	if fields.Contains("md") {
		resp.Metadata = md
	}
	return JSON(c, http.StatusOK, resp)
}
