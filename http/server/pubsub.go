package server

import (
	"io/ioutil"
	"net/http"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/encoding"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"golang.org/x/net/websocket"
)

// TODO: Whitelist publish recipients by default

func (s *Server) publish(c echo.Context) error {
	request := c.Request()
	ctx := request.Context()
	logger.Infof(ctx, "Server POST publish %s", s.urlWithBase(c))

	_, status, err := s.authorize(c)
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

	enc := encoding.MustEncode(bin, encoding.Base64)

	logger.Infof(ctx, "Publish to %s", rid)
	if err := s.mc.Publish(ctx, rid.String(), enc); err != nil {
		return internalError(c, err)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) subscribe(c echo.Context) error {
	request := c.Request()
	ctx := request.Context()

	logger.Infof(ctx, "Server GET subscribe %s", s.urlWithBase(c))

	kid, status, err := s.authorize(c)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		request := ws.Request()
		ctx := request.Context()

		logger.Infof(ctx, "Subscribe %s", kid)
		ch, err := s.mc.Subscribe(ctx, kid.String())
		if err != nil {
			logger.Errorf(ctx, "Error: %v", err)
			return
		}
		defer func() { _ = s.mc.Unsubscribe(ctx, kid.String()) }()

		logger.Infof(ctx, "Waiting for publishes %s", kid)
		for {
			select {
			case b := <-ch:
				dec, err := encoding.Decode(string(b), encoding.Base64)
				if err != nil {
					logger.Errorf(ctx, "Error: %v", err)
					continue
				}

				// logger.Debugf(ctx, "Sending msg: %v", msg)
				if err := websocket.Message.Send(ws, dec); err != nil {
					logger.Errorf(ctx, "Error: %v", err)
					continue
				}
			case <-ctx.Done():
				logger.Errorf(ctx, "Error: %v", ctx.Err())
				return
			}
		}

	}).ServeHTTP(c.Response(), c.Request())
	return nil
}
