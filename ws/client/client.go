package client

import (
	"encoding/json"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/keys-pub/keys-ext/ws/api"
	"github.com/pkg/errors"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 1024 * 64
)

// Client to websocket.
type Client struct {
	url       *url.URL
	conn      *websocket.Conn
	connected bool

	connectMtx sync.Mutex
	writeMtx   sync.Mutex
}

// New creates a websocket client.
func New(urs string) (*Client, error) {
	url, err := url.Parse(urs)
	if err != nil {
		return nil, err
	}
	return &Client{
		url: url,
	}, nil
}

// Authorize tokens.
func (c *Client) Authorize(tokens []string) error {
	if c.conn == nil {
		return errors.Errorf("not connected")
	}
	if len(tokens) == 0 {
		return nil
	}
	if err := c.sendTokens(tokens); err != nil {
		return err
	}
	return nil
}

// Close ...
func (c *Client) Close() {
	logger.Infof("Close...")
	if c.connected {
		err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			// Failed to write close message
			logger.Infof("Failed to write close message: %v", err)
		}
	}
	c.close()
}

func (c *Client) close() {
	if c.conn != nil {
		c.connectMtx.Lock()
		_ = c.conn.Close()
		c.connected = false
		c.connectMtx.Unlock()
	}
}

func (c *Client) connect() error {
	if c.connected {
		return errors.Errorf("already connected")
	}
	logger.Infof("Connect...")
	logger.Infof("Dial %s", c.url)
	conn, _, err := websocket.DefaultDialer.Dial(c.url.String(), nil)
	if err != nil {
		return errors.Wrapf(err, "failed to dial")
	}
	conn.SetReadLimit(maxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error { _ = c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	c.connectMtx.Lock()
	c.conn = conn
	c.connected = true
	c.connectMtx.Unlock()
	return nil
}

// Connect client.
func (c *Client) Connect() error {
	if err := c.connect(); err != nil {
		return err
	}
	return nil
}

// ReadEvents reads events.
func (c *Client) ReadEvents() ([]*api.Event, error) {
	if !c.connected {
		return nil, errors.Errorf("not connected")
	}

	logger.Debugf("Read event")
	_, connMsg, err := c.conn.ReadMessage()
	if err != nil {
		logger.Errorf("Connection error: %v", err)
		c.close()
		return nil, err
	}
	var events []*api.Event
	if err := json.Unmarshal(connMsg, &events); err != nil {
		return nil, err
	}
	return events, nil
}

func (c *Client) sendTokens(tokens []string) error {
	logger.Infof("Sending tokens...")
	c.writeMtx.Lock()
	defer c.writeMtx.Unlock()

	b := []byte(strings.Join(tokens, ","))
	if err := c.conn.WriteMessage(websocket.TextMessage, b); err != nil {
		return errors.Wrapf(err, "failed to write message")
	}
	return nil
}

// Ping sends a ping message.
func (c *Client) Ping() error {
	if c.conn == nil {
		return errors.Errorf("not connected")
	}
	return c.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait))
}
