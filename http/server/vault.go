package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/ds"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) postVault(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	kid, status, err := authorize(c, s.URL, "kid", s.nowFn(), s.rds)
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
	cpath := ds.Path("vaults", kid)
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

	cpath := ds.Path("vaults", kid)
	resp, respErr := s.events(c, cpath)
	if respErr != nil {
		return respErr
	}
	if len(resp.Events) == 0 && resp.Index == 0 {
		return ErrNotFound(c, errors.Errorf("vault not found"))
	}

	return JSON(c, http.StatusOK, resp)
}

func (s *Server) putVault(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	kid, status, err := authorize(c, s.URL, "kid", s.nowFn(), s.rds)
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

	var req []*api.Data
	if err := json.Unmarshal(b, &req); err != nil {
		return ErrBadRequest(c, err)
	}

	ctx := c.Request().Context()
	cpath := ds.Path("vaults", kid)
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
