package server

import (
	"context"
	"log"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/ws/api"
	"github.com/pkg/errors"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	url string

	rds *Redis

	// Registered clients.
	clients map[string]*client

	// Clients by key.
	clientsForKey map[keys.ID]map[string]*client

	// Inbound messages.
	broadcast chan *api.PubEvent

	// Inbound auth.
	auth chan *authClient

	// Register client from the clients.
	register chan *client

	// Unregister requests from clients.
	unregister chan *client

	nonceCheck api.NonceCheck
}

type authClient struct {
	client *client
	kid    keys.ID
}

// NewHub ...
func NewHub(url string) *Hub {
	h := &Hub{
		url:           url,
		broadcast:     make(chan *api.PubEvent),
		auth:          make(chan *authClient, 10),
		register:      make(chan *client),
		unregister:    make(chan *client),
		clients:       make(map[string]*client),
		clientsForKey: make(map[keys.ID]map[string]*client),
	}
	h.nonceCheck = h.rdsNonceCheck
	return h
}

func (h *Hub) rdsNonceCheck(ctx context.Context, nonce string) error {
	if h.rds == nil {
		return errors.Errorf("redis not set")
	}
	if nonce == "" {
		return errors.Errorf("empty nonce")
	}
	val, err := h.rds.Get(ctx, nonce)
	if err != nil {
		return err
	}
	if val != "" {
		return errors.Errorf("nonce collision")
	}
	if err := h.rds.Set(ctx, nonce, "1"); err != nil {
		return err
	}
	if err := h.rds.Expire(ctx, nonce, time.Hour); err != nil {
		return err
	}
	return nil
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
			// log.Printf("register auth %s\n", auth.client.id)
			h.registerAuth(auth)
			log.Printf("send hello %s\n", auth.kid)
			auth.client.send <- &api.Event{Type: api.HelloEvent, User: auth.kid}
		case pevent := <-h.broadcast:
			for _, user := range pevent.Users {
				clients := h.findClients(user)
				event := &api.Event{
					Channel: pevent.Channel,
					User:    user,
					Index:   pevent.Index,
					Type:    api.ChannelEvent,
				}
				for _, client := range clients {
					select {
					case client.send <- event:
						// log.Printf("send %s => %s\n", client.id, message.KID)
					default:
						close(client.send)
						h.unregisterAuth(client)
					}
				}
			}
		}
	}
}

func (h *Hub) registerAuth(auth *authClient) {
	h.registerKIDs(auth.client, auth.kid)
}

func (h *Hub) registerKIDs(cl *client, kid keys.ID) {
	// log.Printf("auth %s => %s\n", auth.client.id, auth.kid)
	if cl.kids == nil {
		cl.kids = []keys.ID{}
	}
	cl.kids = append(cl.kids, kid)
	clients, ok := h.clientsForKey[kid]
	if !ok {
		clients = map[string]*client{}
		h.clientsForKey[kid] = clients
	}
	clients[cl.id] = cl
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
