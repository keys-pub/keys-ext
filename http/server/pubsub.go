package server

import (
	"context"
	"io/ioutil"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/http"
	"github.com/keys-pub/keys/tsutil"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// PubSub ...
type PubSub interface {
	Publish(ctx context.Context, name string, b []byte) error
	Subscribe(ctx context.Context, name string, receiveFn func(b []byte)) error
}

// PubSubServer implements PubSub.
type PubSubServer struct {
	pubSub PubSub
	rds    Redis
	logger Logger
	clock  tsutil.Clock

	URL string
}

// NewPubSubServer creates a PubSubServer.
func NewPubSubServer(pubSub PubSub, rds Redis, logger Logger) *PubSubServer {
	return &PubSubServer{
		pubSub: pubSub,
		rds:    rds,
		logger: logger,
	}
}

// SetClock sets clock.
func (s *PubSubServer) SetClock(clock tsutil.Clock) {
	s.clock = clock
}

// NewPubSubHandler returns http.Handler for Server.
func NewPubSubHandler(s *PubSubServer) http.Handler {
	return newPubSubHandler(s)
}

func newPubSubHandler(s *PubSubServer) *echo.Echo {
	e := echo.New()
	// e.HTTPErrorHandler = s.ErrorHandler
	s.AddRoutes(e)
	return e
}

// AddRoutes adds routes to an Echo instance.
func (s *PubSubServer) AddRoutes(e *echo.Echo) {
	// PubSub
	e.POST("/publish/:kid/:rid", s.publish)
	e.GET("/subscribe/:kid", s.subscribe)

	e.GET("/wsecho", s.wsEcho)
}

// TODO: Whitelist publish recipients by default

func (s *PubSubServer) publish(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	if _, err := s.auth(c, "kid"); err != nil {
		return err
	}

	if c.Request().Body == nil {
		return ErrBadRequest(c, errors.Errorf("missing body"))
	}

	bin, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return s.internalError(c, err)
	}

	if len(bin) > 16*1024 {
		// TODO: Check length before reading data
		return ErrBadRequest(c, errors.Errorf("message too large (greater than 16KiB)"))
	}

	recipient := c.Param("rid")
	if recipient == "" {
		return ErrBadRequest(c, errors.Errorf("no recipient id"))
	}
	rid, err := keys.ParseID(recipient)
	if err != nil {
		return ErrBadRequest(c, err)
	}

	s.logger.Infof("Publish to %s", rid)
	if err := s.pubSub.Publish(ctx, rid.String(), bin); err != nil {
		return s.internalError(c, err)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

func (s *PubSubServer) auth(c echo.Context, param string) (keys.ID, error) {
	return "", ErrForbidden(c, errors.Errorf("not implemented"))
}

func (s *PubSubServer) internalError(c echo.Context, err error) error {
	s.logger.Errorf("Internal error: %v", err)
	return ErrResponse(c, http.StatusInternalServerError, err)
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func (s *PubSubServer) subscribe(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	kid, err := s.auth(c, "kid")
	if err != nil {
		return err
	}

	subCtx, cancel := context.WithCancel(c.Request().Context())
	defer cancel()

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		s.logger.Errorf("Upgrade error: %v", err)
		return ErrBadRequest(c, err)
	}
	defer ws.Close()

	// After connection has been upgraded, don't write to response writer,
	// (write error to websocket and return nil).

	s.logger.Infof("Subscribe %s", kid)

	receiveFn := func(b []byte) {
		if err := ws.WriteMessage(websocket.TextMessage, b); err != nil {
			s.logger.Errorf("Write error: %v", err)
			cancel()
			return
		}
	}

	if err := s.pubSub.Subscribe(subCtx, kid.String(), receiveFn); err != nil {
		s.logger.Errorf("Subscribe error: %v", err)
		return nil
	}

	return nil
}

func (s *PubSubServer) wsEcho(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		s.logger.Errorf("Upgrade error: %v", err)
		return ErrBadRequest(c, err)
	}
	defer ws.Close()

	// After connection has been upgraded, don't write to response writer,
	// (write error to websocket and return nil).

	for {
		typ, msg, err := ws.ReadMessage()
		if err != nil {
			// s.logger.Errorf("Read error: %v", err)
			return nil
		}
		switch typ {
		case websocket.CloseMessage:
			return nil
		case websocket.TextMessage:
			if err := ws.WriteMessage(websocket.TextMessage, msg); err != nil {
				s.logger.Errorf("Write error: %v", err)
				return nil
			}
		}
	}
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
