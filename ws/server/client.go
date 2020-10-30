package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

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
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type client struct {
	hub *Hub

	id string

	// The websocket connection.
	conn *websocket.Conn

	kids []keys.ID

	// Buffered channel of outbound messages.
	send chan *api.Message
}

// newClient returns a new client.
func newClient(hub *Hub, conn *websocket.Conn) *client {
	return &client{
		id:   encoding.MustEncode(keys.Rand16()[:], encoding.Base62),
		hub:  hub,
		conn: conn,
		send: make(chan *api.Message, 256),
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
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("read error: %v", err)
			}
			break
		}
		// log.Printf("read: %s\n", message)

		kid, err := api.CheckAuth(context.TODO(), data, time.Now(), c.hub.host, c.hub.nonces)
		if err != nil {
			log.Printf("auth err: %v\n", err)
			break
		}

		c.hub.auth <- &authClient{client: c, kid: kid}
	}
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
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			b, err := json.Marshal(message)
			if err != nil {
				return
			}
			w.Write(b)

			// Add queued messages .
			n := len(c.send)
			for i := 0; i < n; i++ {
				message = <-c.send
				b, err := json.Marshal(message)
				if err != nil {
					return
				}
				w.Write(b)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Serve handles websocket requests from the peer.
func Serve(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	cl := newClient(hub, conn)
	cl.hub.register <- cl

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go cl.writePump()
	go cl.readPump()
}
