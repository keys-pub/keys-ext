package server

import (
	"net/http"
	"strconv"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/api"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) getUserSearch(c echo.Context) error {
	ctx := c.Request().Context()
	request := c.Request()
	logger.Infof(ctx, "Server GET user search %s", request.URL)

	q := c.QueryParam("q")

	plimit := c.QueryParam("limit")
	if plimit == "" {
		plimit = "100"
	}
	limit, err := strconv.Atoi(plimit)
	if err != nil {
		return ErrBadRequest(c, errors.Wrapf(err, "invalid limit"))
	}

	results, err := s.users.Search(ctx, &keys.UserSearchRequest{Query: q, Limit: limit})
	if err != nil {
		return internalError(c, err)
	}

	users := make([]*api.User, 0, len(results))
	for _, res := range results {
		users = append(users, api.UserFromResult(res.UserResult))
	}

	resp := api.UserSearchResponse{
		Users: users,
	}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) getUser(c echo.Context) error {
	ctx := c.Request().Context()
	request := c.Request()
	logger.Infof(ctx, "Server GET users %s", request.URL)

	kid, err := keys.ParseID(c.Param("kid"))
	if err != nil {
		return ErrNotFound(c, errors.Errorf("kid not found"))
	}

	userResult, err := s.users.Get(ctx, kid)
	if err != nil {
		return internalError(c, err)
	}
	if userResult == nil {
		return ErrNotFound(c, errors.Errorf("user not found"))
	}

	resp := api.UserResponse{
		User: api.UserFromResult(userResult),
	}
	return JSON(c, http.StatusOK, resp)
}
