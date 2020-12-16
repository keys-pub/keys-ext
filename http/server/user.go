package server

import (
	"net/http"
	"strconv"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/users"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) getUserSearch(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	q := c.QueryParam("q")

	plimit := c.QueryParam("limit")
	if plimit == "" {
		plimit = "100"
	}
	limit, err := strconv.Atoi(plimit)
	if err != nil {
		return ErrBadRequest(c, errors.Wrapf(err, "invalid limit"))
	}

	results, err := s.users.Search(ctx, &users.SearchRequest{Query: q, Limit: limit})
	if err != nil {
		return ErrInternalServer(c, err)
	}

	usrs := make([]*api.User, 0, len(results))
	for _, res := range results {
		usrs = append(usrs, api.UserFromSearchResult(res))
	}

	resp := api.UserSearchResponse{
		Users: usrs,
	}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) getUser(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	kid, err := keys.ParseID(c.Param("kid"))
	if err != nil {
		return ErrNotFound(c, errors.Errorf("kid not found"))
	}

	userResult, err := s.users.Find(ctx, kid)
	if err != nil {
		return ErrInternalServer(c, err)
	}
	if userResult == nil {
		return ErrNotFound(c, errors.Errorf("user not found"))
	}

	resp := api.UserResponse{
		User: api.UserFromResult(userResult),
	}
	return JSON(c, http.StatusOK, resp)
}
