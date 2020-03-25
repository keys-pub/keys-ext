package server

import (
	"context"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/keys-pub/keys"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type PubSub interface {
	Publish(ctx context.Context, name string, b []byte) error
	Subscribe(ctx context.Context, name string, receiveFn func(b []byte)) error
}

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

	logger.Infof(ctx, "Publish to %s", rid)
	if err := s.ps.Publish(ctx, rid.String(), bin); err != nil {
		return internalError(c, err)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

var (
	upgrader = websocket.Upgrader{}
)

func (s *Server) subscribe(c echo.Context) error {
	request := c.Request()
	ctx := request.Context()

	logger.Infof(ctx, "Server GET subscribe %s", s.urlWithBase(c))

	kid, status, err := s.authorize(c)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	ctx, cancel := context.WithCancel(request.Context())
	defer cancel()

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	var readErr error
	logger.Infof(ctx, "Subscribe %s", kid)

	receiveFn := func(b []byte) {
		if err := ws.WriteMessage(websocket.TextMessage, b); err != nil {
			readErr = err
			cancel()
			return
		}
	}

	if err := s.ps.Subscribe(ctx, kid.String(), receiveFn); err != nil {
		return internalError(c, err)
	}

	if readErr != nil {
		return readErr
	}

	return nil
}

type pubSub struct {
	sync.Mutex
	subs map[string][][]byte
}

// NewPubSub is PubSub for testing.
func NewPubSub() PubSub {
	return &pubSub{
		subs: map[string][][]byte{},
	}
}

func (p *pubSub) Publish(ctx context.Context, name string, b []byte) error {
	p.Lock()
	defer p.Unlock()
	vals, ok := p.subs[name]
	if ok {
		p.subs[name] = append(vals, b)
	} else {
		p.subs[name] = [][]byte{b}
	}
	return nil
}

func (p *pubSub) Subscribe(ctx context.Context, name string, receiveFn func(b []byte)) error {
	for {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			if err == context.Canceled {
				return nil
			}
			return err
		case <-time.After(time.Millisecond * 10):
			p.Lock()
			vals, ok := p.subs[name]
			delete(p.subs, name)
			if ok {
				for _, v := range vals {
					receiveFn(v)
				}
			}
			p.Unlock()
		}
	}
}
