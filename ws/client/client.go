package client

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/ws/api"
	"github.com/pkg/errors"
)

// Client to websocket.
type Client struct {
	urs  string
	conn *websocket.Conn

	connectMtx sync.Mutex

	keys []*keys.EdX25519Key
}

// New creates a websocket client.
func New(urs string) *Client {
	return &Client{
		urs:  urs,
		keys: []*keys.EdX25519Key{},
	}
}

// Register key.
func (c *Client) Register(key *keys.EdX25519Key) {
	c.connectMtx.Lock()
	defer c.connectMtx.Unlock()

	logger.Infof("register %s", key.ID())
	c.keys = append(c.keys, key)
	conn := c.conn
	if conn != nil {
		if err := sendAuth(conn, c.urs, key); err != nil {
			c.close()
		}
	}
}

// Close ...
func (c *Client) Close(sendClose bool) {
	logger.Infof("close")
	c.connectMtx.Lock()
	defer c.connectMtx.Unlock()

	if c.conn != nil {
		if sendClose {
			err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				// Failed to write close message
			}
		}
		c.close()
	}
}

func (c *Client) close() {
	c.conn.Close()
	c.conn = nil
}

// Connect client.
func (c *Client) Connect() error {
	logger.Infof("connect")
	c.connectMtx.Lock()
	defer c.connectMtx.Unlock()

	if c.conn != nil {
		return errors.Errorf("already connected")
	}
	logger.Infof("dial %s", c.urs)
	conn, _, err := websocket.DefaultDialer.Dial(c.urs, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to dial")
	}

	for _, key := range c.keys {
		if err := sendAuth(conn, c.urs, key); err != nil {
			c.close()
			return errors.Wrapf(err, "failed to send auth")
		}
	}

	c.conn = conn
	return nil
}

// ReadMessage reads a message.
func (c *Client) ReadMessage() (*api.Message, error) {
	if c.conn == nil {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	logger.Infof("read message")
	_, message, err := c.conn.ReadMessage()
	if err != nil {
		c.close()
		return nil, err
	}
	var msg api.Message
	if err := json.Unmarshal(message, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func sendAuth(conn *websocket.Conn, urs string, key *keys.EdX25519Key) error {
	logger.Infof("send auth %s", key.ID())
	b := api.GenerateAuth(key, urs)
	if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
		return errors.Wrapf(err, "failed to write message")
	}
	return nil
}
