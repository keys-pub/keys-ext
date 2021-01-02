package server

import (
	"encoding/json"
	"net/http/httptest"

	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/http"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// Experimental, not enabled.
func (s *Server) postBatch(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	body, st, err := readBody(c, false, 64*1024)
	if err != nil {
		return s.ErrResponse(c, st, err)
	}
	var req api.BatchRequests
	if err := json.Unmarshal(body, &req); err != nil {
		return s.ErrBadRequest(c, errors.Errorf("invalid request"))
	}
	if len(req.Requests) > 100 {
		return s.ErrBadRequest(c, errors.Errorf("invalid request"))
	}

	responses := []*api.BatchResponse{}
	for _, req := range req.Requests {
		hreq, err := http.NewRequest(req.Method, req.URL, nil)
		if err != nil {
			responses = append(responses, &api.BatchResponse{
				ID:     req.ID,
				Status: http.StatusBadRequest,
			})
			continue
		}
		for hkey, hval := range req.Headers {
			hreq.Header.Add(hkey, hval)
		}

		// TODO: Use buffer pool or more effecient http.ResponseWriter?
		resp := httptest.NewRecorder()

		c.Echo().ServeHTTP(resp, hreq)

		var val interface{}
		if err := json.Unmarshal(resp.Body.Bytes(), &val); err != nil {
			responses = append(responses, &api.BatchResponse{
				ID:     req.ID,
				Status: http.StatusBadRequest,
			})
		}

		responses = append(responses, &api.BatchResponse{
			ID:     req.ID,
			Status: resp.Code,
			Body:   val,
		})
	}

	out := api.BatchResponses{
		Responses: responses,
	}
	return c.JSON(http.StatusOK, out)
}
