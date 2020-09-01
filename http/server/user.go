package server

import (
	"net/http"
	"strconv"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/user"
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

	results, err := s.users.Search(ctx, &user.SearchRequest{Query: q, Limit: limit})
	if err != nil {
		return s.internalError(c, err)
	}

	users := make([]*api.User, 0, len(results))
	for _, res := range results {
		users = append(users, api.UserFromSearchResult(res))
	}

	resp := api.UserSearchResponse{
		Users: users,
	}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) getUsers(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	kid, err := keys.ParseID(c.Param("kid"))
	if err != nil {
		return ErrNotFound(c, errors.Errorf("invalid kid"))
	}

	results, err := s.users.Find(ctx, kid)
	if err != nil {
		return s.internalError(c, err)
	}
	if len(results) == 0 {
		return ErrNotFound(c, errors.Errorf("no users found"))
	}

	resp := api.UsersResponse{
		Users: api.UsersFromResults(results),
	}
	return JSON(c, http.StatusOK, resp)
}

// getUser is deprecated
func (s *Server) getUser(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	kid, err := keys.ParseID(c.Param("kid"))
	if err != nil {
		return ErrNotFound(c, errors.Errorf("invalid kid"))
	}

	results, err := s.users.Find(ctx, kid)
	if err != nil {
		return s.internalError(c, err)
	}
	if len(results) == 0 {
		return ErrNotFound(c, errors.Errorf("no users found"))
	}

	resp := struct {
		User *api.User `json:"user"`
	}{
		User: api.UserFromResult(results[0]),
	}
	return JSON(c, http.StatusOK, resp)
}
