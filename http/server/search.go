package server

import (
	"net/http"
	"strconv"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/api"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) getSearch(c echo.Context) error {
	ctx := c.Request().Context()
	request := c.Request()
	logger.Infof(ctx, "Server GET search %s", request.URL)

	q := c.QueryParam("q")

	pindex := c.QueryParam("index")
	if pindex == "" {
		pindex = "0"
	}
	index, err := strconv.Atoi(pindex)
	if err != nil {
		return ErrBadRequest(c, errors.Wrapf(err, "invalid index"))
	}

	plimit := c.QueryParam("limit")
	if plimit == "" {
		plimit = "100"
	}
	limit, err := strconv.Atoi(plimit)
	if err != nil {
		return ErrBadRequest(c, errors.Wrapf(err, "invalid limit"))
	}

	// cat := c.QueryParam("cat")
	// cats := keys.NewStringSetSplit(cat, ",")

	// if cats.Size() == 0 || cats.Contains("user") {
	results, err := s.search.Search(ctx, &keys.SearchRequest{Query: q, Index: index, Limit: limit, KIDs: true})
	if err != nil {
		return internalError(c, err)
	}

	resp := api.SearchResponse{
		Results: results,
	}
	return JSON(c, http.StatusOK, resp)
}
