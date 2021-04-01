package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/websocket"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/ws/api"
	"github.com/keys-pub/keys/encoding"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024 * 16
)

var upgrader = websocket.Upgrader{
	// ReadBufferSize:  1024,
	// WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type client struct {
	hub *Hub

	id string

	// The websocket connection.
	conn *websocket.Conn

	tokens   []string
	tokenKey *[32]byte

	// Buffered channel of outbound messages.
	send chan *api.Event
}

// newClient returns a new client.
func newClient(hub *Hub, conn *websocket.Conn, tokenKey *[32]byte) *client {
	return &client{
		id:       encoding.MustEncode(keys.Rand16()[:], encoding.Base62),
		hub:      hub,
		conn:     conn,
		send:     make(chan *api.Event, 256),
		tokenKey: tokenKey,
	}
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *client) readPump() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { _ = c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("read error: %v\n", err)
			}
			return
		}

		tokens := strings.Split(string(data), ",")
		for _, token := range tokens {
			t, err := jwt.Parse(token, c.jwtTokens)
			if err != nil {
				log.Printf("invalid token\n")
				return
			}
			if err := t.Claims.Valid(); err != nil {
				log.Printf("invalid token (claims)\n")
				return
			}

			c.hub.auth <- &authClient{client: c, token: token}
		}
	}
}

func (c *client) jwtTokens(t *jwt.Token) (interface{}, error) {
	return c.tokenKey[:], nil
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case event, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			b, err := json.Marshal(event)
			if err != nil {
				return
			}
			if _, err = w.Write([]byte("[")); err != nil {
				return
			}
			if _, err = w.Write(b); err != nil {
				return
			}

			// Add queued messages.
			n := len(c.send)
			for i := 0; i < n; i++ {
				event = <-c.send
				b, err := json.Marshal(event)
				if err != nil {
					return
				}
				if _, err = w.Write([]byte(",")); err != nil {
					return
				}
				if _, err = w.Write(b); err != nil {
					return
				}
			}
			if _, err = w.Write([]byte("]")); err != nil {
				return
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Serve handles websocket requests from the peer.
func Serve(hub *Hub, tokenKey *[32]byte, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	cl := newClient(hub, conn, tokenKey)
	cl.hub.register <- cl

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go cl.writePump()
	go cl.readPump()
}
