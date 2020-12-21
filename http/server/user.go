package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/user"
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
		return s.ErrBadRequest(c, errors.Wrapf(err, "invalid limit"))
	}

	results, err := s.users.Search(ctx, &users.SearchRequest{Query: q, Limit: limit})
	if err != nil {
		return s.ErrInternalServer(c, err)
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

	param := strings.TrimSpace(c.Param("user"))

	var userResult *user.Result
	if strings.Contains(param, "@") {
		ur, err := s.users.User(ctx, param)
		if err != nil {
			return s.ErrInternalServer(c, err)
		}
		userResult = ur
	} else {
		kid, err := keys.ParseID(param)
		if err != nil {
			return s.ErrNotFound(c, errors.Errorf("user not found"))
		}
		ur, err := s.users.Find(ctx, kid)
		if err != nil {
			return s.ErrInternalServer(c, err)
		}
		userResult = ur
	}

	if userResult == nil {
		return s.ErrNotFound(c, errors.Errorf("user not found"))
	}

	resp := api.UserResponse{
		User: api.UserFromResult(userResult),
	}
	return JSON(c, http.StatusOK, resp)
}
