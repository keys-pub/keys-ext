package api

import (
	"time"

	"github.com/keys-pub/keys"
)

// Response ...
type Response struct {
	Error *Error `json:"error,omitempty"`
}

// Error ...
type Error struct {
	Message string `json:"message,omitempty"`
	Status  int    `json:"status,omitempty"`
}

// Metadata ...
type Metadata struct {
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// SigchainResponse is the response format for a Sigchain request.
type SigchainResponse struct {
	KID        keys.ID            `json:"kid"`
	Metadata   map[string]Metadata `json:"md,omitempty"`
	Statements []*keys.Statement  `json:"statements"`
}

// MetadataFor returns metadata for Signed.
func (r SigchainResponse) MetadataFor(st *keys.Statement) Metadata {
	md, ok := r.Metadata[st.URLPath()]
	if !ok {
		return Metadata{}
	}
	return md
}

// Sigchain from response.
func (r SigchainResponse) Sigchain() (*keys.Sigchain, error) {
	spk, err := keys.DecodeSignPublicKey(r.KID.String())
	if err != nil {
		return nil, err
	}
	sc := keys.NewSigchain(spk)
	for _, st := range r.Statements {
		// md := r.MetadataFor(st)
		// if md.CreatedAt.IsZero() {
		// 	return nil, errors.Errorf("missing metadata for statement in response")
		// }
		if err := sc.Add(st); err != nil {
			return nil, err
		}
	}
	return sc, nil
}

// SigchainsResponse is the response format for a listing all sigchain
// statements.
type SigchainsResponse struct {
	Metadata   map[string]Metadata `json:"md,omitempty"`
	Statements []*keys.Statement  `json:"statements"`
	Version    string              `json:"version"`
}

// MetadataFor returns metadata for Signed.
func (r SigchainsResponse) MetadataFor(st *keys.Statement) Metadata {
	md, ok := r.Metadata[st.URLPath()]
	if !ok {
		return Metadata{}
	}
	return md
}

// Message ...
type Message struct {
	Data []byte   `json:"data"`
	ID   keys.ID `json:"id"`
	Path string   `json:"path"`
}

// MessagesResponse is the response from messages.
type MessagesResponse struct {
	KID      keys.ID            `json:"kid"`
	Messages []*Message          `json:"messages"`
	Metadata map[string]Metadata `json:"md,omitempty"`
	Version  string              `json:"version"`
}

// MetadataFor returns metadata for Message.
func (r MessagesResponse) MetadataFor(msg *Message) Metadata {
	md, ok := r.Metadata[msg.Path]
	if !ok {
		return Metadata{}
	}
	return md
}

// SearchResponse ...
type SearchResponse struct {
	Results []*keys.SearchResult `json:"results"`
}

// Item ...
type Item struct {
	Data []byte   `json:"data"`
	ID   keys.ID `json:"id"`
	Path string   `json:"path"`
}

// VaultResponse ...
type VaultResponse struct {
	KID      keys.ID            `json:"kid"`
	Items    []*Item             `json:"items"`
	Metadata map[string]Metadata `json:"md,omitempty"`
	Version  string              `json:"version"`
}

// MetadataFor returns metadata for Signed.
func (r VaultResponse) MetadataFor(item *Item) Metadata {
	md, ok := r.Metadata[item.Path]
	if !ok {
		return Metadata{}
	}
	return md
}
