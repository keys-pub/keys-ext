package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/keys-pub/keys"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"golang.org/x/net/websocket"
)

func (s *Server) publish(c echo.Context) error {
	request := c.Request()
	ctx := request.Context()
	logger.Infof(ctx, "Server POST publish %s", s.urlString(c))

	kid, status, err := s.authorize(c)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	recipient := c.Param("rid")
	if recipient == "" {
		return ErrBadRequest(c, errors.Errorf("no recipient id"))
	}
	rid, err := keys.ParseID(recipient)
	if err != nil {
		return ErrBadRequest(c, err)
	}

	if c.Request().Body == nil {
		return ErrBadRequest(c, errors.Errorf("missing body"))
	}

	bin, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return internalError(c, err)
	}

	if len(bin) > 16*1024 {
		// TODO: Check length before reading data
		return ErrBadRequest(c, errors.Errorf("message too large (greater than 16KiB)"))
	}

	verified, err := verifyStatement(bin)
	if err != nil {
		return ErrBadRequest(c, err)
	}
	if verified.KID != kid {
		return ErrBadRequest(c, errors.Errorf("statement kid mismatch"))
	}

	logger.Infof(ctx, "Publish to %s", rid)
	if err := s.mc.Publish(ctx, rid.String(), string(bin)); err != nil {
		return internalError(c, err)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

func verifyStatement(b []byte) (*keys.Statement, error) {
	var st keys.Statement
	if err := json.Unmarshal(b, &st); err != nil {
		return nil, errors.Errorf("not a statement")
	}

	if err := st.Verify(); err != nil {
		return nil, errors.Errorf("statement failed to verify")
	}

	return &st, nil
}

func (s *Server) subscribe(c echo.Context) error {
	request := c.Request()
	ctx := request.Context()
	logger.Infof(ctx, "Server GET subscribe %s", s.urlString(c))

	kid, status, err := s.authorize(c)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		request := ws.Request()
		ctx := request.Context()

		ch := make(chan []byte)
		logger.Infof(ctx, "Subscribe %s", kid)
		if err := s.mc.Subscribe(ctx, kid.String(), ch); err != nil {
			logger.Errorf(ctx, "Error: %v", err)
			return
		}

		logger.Infof(ctx, "Waiting for publishes %s", kid)
		for {
			select {
			case b := <-ch:
				verified, err := verifyStatement(b)
				if err != nil {
					logger.Errorf(ctx, "Invalid statement in memcache: %v", err)
				}

				// logger.Debugf(ctx, "Sending msg: %v", msg)
				if err := websocket.JSON.Send(ws, verified); err != nil {
					logger.Errorf(ctx, "Error: %v", err)
				}
			case <-ctx.Done():
				logger.Errorf(ctx, "Error: %v", ctx.Err())
				return
			}
		}

	}).ServeHTTP(c.Response(), c.Request())
	return nil
}
