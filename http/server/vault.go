package server

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/ds"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) postVault(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	kid, status, err := authorize(c, s.URL, "kid", s.nowFn(), s.mc)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	var data []byte
	if c.Request().Body != nil {
		b, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return s.internalError(c, err)
		}

		if len(b) > 16*1024 {
			// TODO: Check length before reading data
			return ErrBadRequest(c, errors.Errorf("message too large (greater than 16KiB)"))
		}
		data = b
	}

	ctx := c.Request().Context()

	// path := ds.Path("vaults", kid, "items", id)
	// if err := s.fi.Set(ctx, path, mb); err != nil {
	// 	return s.internalError(c, err)
	// }

	cpath := ds.Path("vaults", kid, "changes")
	if _, err := s.fi.ChangeAdd(ctx, cpath, data); err != nil {
		return s.internalError(c, err)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) listVault(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	kid, status, err := authorize(c, s.URL, "kid", s.nowFn(), s.mc)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	cpath := ds.Path("vaults", kid, "changes")
	chgs, clientErr, err := s.changes(c, cpath)
	if err != nil {
		return s.internalError(c, err)
	}
	if clientErr != nil {
		return ErrResponse(c, http.StatusBadRequest, clientErr.Error())
	}
	if len(chgs.changes) == 0 && chgs.version == 0 {
		return ErrNotFound(c, errors.Errorf("vault not found"))
	}

	items := make([]*api.VaultItem, 0, len(chgs.changes))
	for _, chg := range chgs.changes {
		item, err := s.vaultItemFromChange(chg)
		if err != nil {
			return s.internalError(c, err)
		}
		if item == nil {
			continue
		}
		items = append(items, item)
	}

	resp := api.VaultResponse{
		Items:   items,
		Version: fmt.Sprintf("%d", chgs.versionNext),
	}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) vaultItemFromChange(chg *ds.Change) (*api.VaultItem, error) {
	if chg == nil {
		return nil, nil
	}
	return &api.VaultItem{
		Data:      chg.Data,
		Timestamp: chg.Timestamp,
	}, nil
}
