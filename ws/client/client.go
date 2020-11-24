package client

import (
	"encoding/json"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/ws/api"
	"github.com/pkg/errors"
)

// Client to websocket.
type Client struct {
	url       *url.URL
	conn      *websocket.Conn
	connected bool

	connectMtx sync.Mutex
	writeMtx   sync.Mutex

	keys []*keys.EdX25519Key
}

// New creates a websocket client.
func New(urs string) (*Client, error) {
	url, err := url.Parse(urs)
	if err != nil {
		return nil, err
	}
	return &Client{
		url:  url,
		keys: []*keys.EdX25519Key{},
	}, nil
}

// Authorize with key.
func (c *Client) Authorize(key *keys.EdX25519Key) {
	// logger.Infof("auth %s", key.ID())
	c.keys = append(c.keys, key)
	if c.connected {
		if err := c.sendAuth(key); err != nil {
			c.close()
		}
	}
}

// Close ...
func (c *Client) Close() {
	logger.Infof("Close...")
	if c.connected {
		err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			// Failed to write close message
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
	logger.Infof("Connect...")
	if c.connected {
		return errors.Errorf("already connected")
	}
	logger.Infof("Dial %s", c.url)
	conn, _, err := websocket.DefaultDialer.Dial(c.url.String(), nil)
	if err != nil {
		return errors.Wrapf(err, "failed to dial")
	}
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

	for _, key := range c.keys {
		if err := c.sendAuth(key); err != nil {
			return errors.Wrapf(err, "failed to send auth")
		}
	}

	return nil
}

// ReadEvents reads events.
func (c *Client) ReadEvents() ([]*api.Event, error) {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return nil, err
		}
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

func (c *Client) sendAuth(key *keys.EdX25519Key) error {
	c.writeMtx.Lock()
	defer c.writeMtx.Unlock()

	logger.Infof("Send auth %s", key.ID())
	b := api.GenerateAuth(key, c.url.String())

	if err := c.conn.WriteMessage(websocket.TextMessage, b); err != nil {
		return errors.Wrapf(err, "failed to write message")
	}
	return nil
}
