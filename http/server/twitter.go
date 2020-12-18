package server

import (
	"fmt"

	"github.com/keys-pub/keys/http"
	"github.com/keys-pub/keys/user/services"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) checkTwitter(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.Param("name")
	if name == "" {
		return s.ErrBadRequest(c, errors.Errorf("invalid name"))
	}
	id := c.Param("id")
	if id == "" {
		return s.ErrBadRequest(c, errors.Errorf("invalid id"))
	}

	twitter := services.Twitter
	api, err := twitter.ValidateURL(name, fmt.Sprintf("https://twitter.com/%s/status/%s", name, id))
	if err != nil {
		return s.ErrBadRequest(c, errors.Errorf("invalid request"))
	}

	// TODO: Rate limit
	body, err := twitter.Request(ctx, s.client, api)
	if err != nil {
		return s.ErrBadRequest(c, errors.Errorf("twitter request failed"))
	}
	if body == nil {
		return s.ErrNotFound(c, errors.Errorf("tweet not found"))
	}

	msg, err := twitter.CheckContent(name, body)
	if err != nil {
		return s.ErrNotFound(c, nil)
	}

	return c.Blob(http.StatusOK, echo.MIMEOctetStream, []byte(msg))
}
