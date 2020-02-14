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
	KID        keys.ID             `json:"kid"`
	Metadata   map[string]Metadata `json:"md,omitempty"`
	Statements []*keys.Statement   `json:"statements"`
}

// MetadataFor returns metadata for Signed.
func (r SigchainResponse) MetadataFor(st *keys.Statement) Metadata {
	md, ok := r.Metadata[st.URL()]
	if !ok {
		return Metadata{}
	}
	return md
}

// Sigchain from response.
func (r SigchainResponse) Sigchain() (*keys.Sigchain, error) {
	spk, err := keys.SigchainPublicKeyFromID(r.KID)
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
	Statements []*keys.Statement   `json:"statements"`
	Version    string              `json:"version"`
}

// MetadataFor returns metadata for Signed.
func (r SigchainsResponse) MetadataFor(st *keys.Statement) Metadata {
	md, ok := r.Metadata[st.URL()]
	if !ok {
		return Metadata{}
	}
	return md
}

// UserFromResult returns User from keys.UserResult.
func UserFromResult(result *keys.UserResult) *User {
	if result == nil {
		return nil
	}
	return &User{
		ID:         result.User.Name + "@" + result.User.Service,
		KID:        result.User.KID.String(),
		Seq:        int32(result.User.Seq),
		Service:    result.User.Service,
		Name:       result.User.Name,
		URL:        result.User.URL,
		Status:     result.Status,
		VerifiedAt: int64(result.VerifiedAt),
		Err:        result.Err,
	}
}

// User ...
type User struct {
	ID         string          `json:"id,omitempty"`
	Name       string          `json:"name,omitempty"`
	KID        string          `json:"kid,omitempty"`
	Seq        int32           `json:"seq,omitempty"`
	Service    string          `json:"service,omitempty"`
	URL        string          `json:"url,omitempty"`
	Status     keys.UserStatus `json:"status,omitempty"`
	VerifiedAt int64           `json:"verifiedAt,omitempty"`
	Err        string          `json:"err,omitempty"`
}

// UserResponse ...
type UserResponse struct {
	User *User `json:"user"`
}

// UserSearchResponse ...
type UserSearchResponse struct {
	Users []*User `json:"users"`
}

// Message ...
type Message struct {
	Data []byte `json:"data"`
	ID   string `json:"id"`
	Path string `json:"path"`
}

// MessagesResponse is the response from messages.
type MessagesResponse struct {
	KID      keys.ID             `json:"kid"`
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
