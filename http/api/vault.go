package api

import "github.com/keys-pub/keys/dstore/events"

// VaultResponse ...
type VaultResponse struct {
	Vault     []*events.Event `json:"vault" msgpack:"vault"`
	Index     int64           `json:"idx" msgpack:"idx"`
	Truncated bool            `json:"truncated,omitempty" msgpack:"trunc,omitempty"`
}
