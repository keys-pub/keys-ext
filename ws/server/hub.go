package server

import (
	"log"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/ws/api"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	host   string
	nonces api.Nonces

	// Registered clients.
	clients map[string]*client

	// Clients by key.
	clientsForKey map[keys.ID]map[string]*client

	// Inbound messages.
	broadcast chan *api.Message

	// Inbound auth.
	auth chan *authClient

	// Register client from the clients.
	register chan *client

	// Unregister requests from clients.
	unregister chan *client
}

type authClient struct {
	client *client
	kid    keys.ID
}

// NewHub ...
func NewHub(host string) *Hub {
	return &Hub{
		host:          host,
		broadcast:     make(chan *api.Message),
		auth:          make(chan *authClient),
		register:      make(chan *client),
		unregister:    make(chan *client),
		clients:       make(map[string]*client),
		clientsForKey: make(map[keys.ID]map[string]*client),
	}
}

// Run hub.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			log.Printf("register %s\n", client.id)
			h.clients[client.id] = client
		case client := <-h.unregister:
			log.Printf("unregister %s\n", client.id)
			if _, ok := h.clients[client.id]; ok {
				h.unregisterAuth(client)
				close(client.send)
			}
		case auth := <-h.auth:
			h.registerAuth(auth)
			auth.client.send <- &api.Message{KID: auth.kid, Type: api.Hello}
		case message := <-h.broadcast:
			clients := h.findClients(message.KID)
			for _, client := range clients {
				select {
				case client.send <- message:
					// log.Printf("send %s => %s\n", client.id, message.KID)
				default:
					close(client.send)
					h.unregisterAuth(client)
				}
			}
		}
	}
}

func (h *Hub) registerAuth(auth *authClient) {
	// log.Printf("auth %s => %s\n", auth.client.id, auth.kid)
	if auth.client.kids == nil {
		auth.client.kids = []keys.ID{}
	}
	auth.client.kids = append(auth.client.kids, auth.kid)
	clients, ok := h.clientsForKey[auth.kid]
	if !ok {
		clients = map[string]*client{}
		h.clientsForKey[auth.kid] = clients
	}
	clients[auth.client.id] = auth.client
}

func (h *Hub) unregisterAuth(cl *client) {
	for _, kid := range cl.kids {
		clientsForKey, ok := h.clientsForKey[kid]
		if !ok {
			continue
		}
		// log.Printf("deauth %s => %s\n", cl.id, kid)
		if _, ok := clientsForKey[cl.id]; ok {
			delete(clientsForKey, cl.id)
		}
	}
	delete(h.clients, cl.id)
	cl.kids = nil
}

func (h *Hub) findClients(kid keys.ID) []*client {
	clients, ok := h.clientsForKey[kid]
	if !ok {
		return nil
	}
	out := make([]*client, 0, len(clients))
	for _, cl := range clients {
		if _, ok := h.clients[cl.id]; !ok {
			continue
		}
		out = append(out, cl)
	}
	return out
}
