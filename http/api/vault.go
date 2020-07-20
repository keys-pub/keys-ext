package api

import "github.com/keys-pub/keys/docs/events"

// VaultResponse ...
type VaultResponse struct {
	Vault []*events.Event `json:"vault"`
	Index int64           `json:"idx"`
}
