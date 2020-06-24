package server

import (
	"encoding/json"
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

	// TODO: max vault size

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
	cpath := ds.Path("vaults", kid, "changes")
	if err := s.fi.ChangesAdd(ctx, cpath, [][]byte{data}); err != nil {
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

	boxes := make([]*api.VaultBox, 0, len(chgs.changes))
	for _, chg := range chgs.changes {
		box, err := s.vaultBoxFromChange(chg)
		if err != nil {
			return s.internalError(c, err)
		}
		if box == nil {
			continue
		}
		boxes = append(boxes, box)
	}

	resp := api.VaultResponse{
		Boxes:   boxes,
		Version: fmt.Sprintf("%d", chgs.versionNext),
	}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) vaultBoxFromChange(chg *ds.Change) (*api.VaultBox, error) {
	if chg == nil {
		return nil, nil
	}
	return &api.VaultBox{
		Data:      chg.Data,
		Version:   chg.Version,
		Timestamp: chg.Timestamp,
	}, nil
}

func (s *Server) putVault(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	kid, status, err := authorize(c, s.URL, "kid", s.nowFn(), s.mc)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	// TODO: max vault size

	if c.Request().Body == nil {
		return ErrBadRequest(c, errors.Errorf("no body data"))
	}
	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return s.internalError(c, err)
	}

	if len(b) > 1024*1024 {
		// TODO: Check length before reading data
		return ErrBadRequest(c, errors.Errorf("message too large (greater than 1MiB)"))
	}

	var vault []*api.VaultBox
	if err := json.Unmarshal(b, &vault); err != nil {
		return ErrBadRequest(c, err)
	}

	ctx := c.Request().Context()
	cpath := ds.Path("vaults", kid, "changes")
	data := make([][]byte, 0, len(vault))
	for _, v := range vault {
		data = append(data, v.Data)
	}
	if err := s.fi.ChangesAdd(ctx, cpath, data); err != nil {
		return s.internalError(c, err)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}
