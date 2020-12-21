package server

import (
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/http"
	"github.com/keys-pub/keys/user"
	"github.com/keys-pub/keys/user/services"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) checkTwitter(c echo.Context) error {
	ctx := c.Request().Context()

	kid, err := keys.ParseID(c.Param("kid"))
	if err != nil {
		return s.ErrBadRequest(c, errors.Errorf("invalid kid"))
	}
	name := c.Param("name")
	if name == "" {
		return s.ErrBadRequest(c, errors.Errorf("invalid name"))
	}
	id := c.Param("id")
	if id == "" {
		return s.ErrBadRequest(c, errors.Errorf("invalid id"))
	}

	twitter := services.Twitter

	urs := "https://twitter.com/" + name + "/status/" + id
	usr, err := user.New(kid, "twitter", name, urs, 1)
	if err != nil {
		return s.ErrBadRequest(c, err)
	}

	// TODO: Rate limit
	_, body, err := twitter.Request(ctx, s.client, usr)
	if err != nil {
		return s.ErrBadRequest(c, errors.Errorf("twitter request failed"))
	}
	if body == nil {
		return s.ErrNotFound(c, errors.Errorf("tweet not found"))
	}

	_, msg, err := twitter.Verify(ctx, body, usr)
	if err != nil {
		return s.ErrNotFound(c, nil)
	}

	return c.Blob(http.StatusOK, echo.MIMEOctetStream, []byte(msg))
}
