package server

import (
	"net/http"
	"strconv"

	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/ds"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) events(c echo.Context, path string) (*api.EventsResponse, error) {
	request := c.Request()
	ctx := request.Context()

	var index int64
	if f := c.QueryParam("idx"); f != "" {
		i, err := strconv.Atoi(f)
		if err != nil {
			return nil, ErrResponse(c, http.StatusBadRequest, errors.Wrapf(err, "invalid index").Error())
		}
		index = int64(i)
	}
	plimit := c.QueryParam("limit")
	if plimit == "" {
		plimit = "100"
	}
	limit, err := strconv.Atoi(plimit)
	if err != nil {
		return nil, ErrResponse(c, http.StatusBadRequest, errors.Wrapf(err, "invalid limit").Error())
	}
	if limit > 100 {
		return nil, ErrResponse(c, http.StatusBadRequest, "invalid limit, too large")
	}

	pdir := c.QueryParam("dir")
	if pdir == "" {
		pdir = "asc"
	}

	var dir ds.Direction
	switch pdir {
	case "asc":
		dir = ds.Ascending
	case "desc":
		dir = ds.Descending
	default:
		return nil, ErrResponse(c, http.StatusBadRequest, "invalid dir")
	}

	s.logger.Infof("Events %s (from=%d)", path, index)
	iter, err := s.fi.Events(ctx, path, index, limit, dir)
	if err != nil {
		return nil, s.internalError(c, err)
	}
	defer iter.Release()
	to := int64(index)
	events := []*ds.Event{}
	for {
		event, err := iter.Next()
		if err != nil {
			return nil, s.internalError(c, err)
		}
		if event == nil {
			break
		}
		events = append(events, event)
		to = event.Index
	}
	s.logger.Infof("Events %s, got %d, (to=%d)", path, len(events), to)

	return &api.EventsResponse{
		Events: events,
		Index:  to,
	}, nil
}
