package server

import (
	"log"

	"github.com/keys-pub/keys-ext/ws/api"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	url string

	rds *Redis

	// Registered clients.
	clients map[string]*client

	// Clients by token.
	clientsByToken map[string]map[string]*client

	// Inbound messages.
	broadcast chan *api.Event

	// Inbound auth.
	auth chan *authClient

	// Register client from the clients.
	register chan *client

	// Unregister requests from clients.
	unregister chan *client
}

type authClient struct {
	client *client
	token  string
}

// NewHub ...
func NewHub(url string) *Hub {
	h := &Hub{
		url:            url,
		broadcast:      make(chan *api.Event),
		auth:           make(chan *authClient, 10),
		register:       make(chan *client),
		unregister:     make(chan *client),
		clients:        make(map[string]*client),
		clientsByToken: make(map[string]map[string]*client),
	}
	return h
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
			// log.Printf("send hello %s\n", auth.kid)
			// auth.client.send <- &api.Event{}
		case event := <-h.broadcast:
			clients := h.findClients(event.Token)
			h.send(clients, event)
		}
	}
}

func (h *Hub) send(clients []*client, event *api.Event) {
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

func (h *Hub) registerAuth(auth *authClient) {
	h.registerClient(auth.client, auth.token)
}

func (h *Hub) registerClient(cl *client, token string) {
	// log.Printf("auth %s => %s\n", auth.client.id, auth.kid)
	if cl.tokens == nil {
		cl.tokens = []string{}
	}
	cl.tokens = append(cl.tokens, token)
	clients, ok := h.clientsByToken[token]
	if !ok {
		clients = map[string]*client{}
		h.clientsByToken[token] = clients
	}
	clients[cl.id] = cl
}

func (h *Hub) unregisterAuth(cl *client) {
	for _, token := range cl.tokens {
		clientsByToken, ok := h.clientsByToken[token]
		if !ok {
			continue
		}
		// log.Printf("deauth %s => %s\n", cl.id, kid)
		delete(clientsByToken, cl.id)
	}
	delete(h.clients, cl.id)
	cl.tokens = nil
}

func (h *Hub) findClients(token string) []*client {
	clients, ok := h.clientsByToken[token]
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

/*
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
*/
