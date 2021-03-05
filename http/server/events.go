package server

import (
	"net/http"

	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore/events"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) events(c echo.Context, path string, max int) (*api.EventsResponse, error) {
	request := c.Request()
	ctx := request.Context()

	index, err := queryParamInt(c, "idx", 0)
	if err != nil {
		return nil, newError(http.StatusBadRequest, err)
	}

	limit, err := queryParamInt(c, "limit", 0)
	if err != nil {
		return nil, newError(http.StatusBadRequest, err)
	}

	if limit == 0 || limit > max {
		limit = max
	}

	pdir := c.QueryParam("order")
	if pdir == "" {
		pdir = "asc"
	}

	var dir events.Direction
	switch pdir {
	case "asc":
		dir = events.Ascending
	case "desc":
		dir = events.Descending
	default:
		return nil, newError(http.StatusBadRequest, errors.Errorf("invalid order"))
	}

	s.logger.Infof("Events %s (from=%d)", path, index)
	iter, err := s.fi.Events(ctx, path, events.Index(int64(index)), events.Limit(int64(limit)), events.WithDirection(dir))
	if err != nil {
		return nil, newError(http.StatusInternalServerError, err)
	}
	defer iter.Release()
	to := int64(index)
	events := []*events.Event{}
	for {
		event, err := iter.Next()
		if err != nil {
			return nil, newError(http.StatusInternalServerError, err)
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
