package api

import (
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/dstore/events"
)

// VaultResponse ...
type VaultResponse struct {
	Vault     []*events.Event `json:"vault" msgpack:"vault"`
	Index     int64           `json:"idx" msgpack:"idx"`
	Truncated bool            `json:"truncated,omitempty" msgpack:"trunc,omitempty"`
}

// VaultStatus ...
type VaultStatus struct {
	ID        keys.ID `json:"id" msgpack:"id"`
	Index     int64   `json:"idx" msgpack:"idx"`
	Timestamp int64   `json:"ts" msgpack:"ts"`
}

// VaultsStatusRequest ...
type VaultsStatusRequest struct {
	Vaults map[keys.ID]string `json:"vaults,omitempty" msgpack:"vaults,omitempty"`
}

// VaultsStatusResponse ...
type VaultsStatusResponse struct {
	Vaults []*VaultStatus `json:"vaults,omitempty" msgpack:"vaults,omitempty"`
}

// VaultToken ...
type VaultToken struct {
	KID   keys.ID `json:"kid"`
	Token string  `json:"token"`
}
