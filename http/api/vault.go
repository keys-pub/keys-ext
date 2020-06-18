package api

import "time"

// VaultItem ...
type VaultItem struct {
	Data      []byte    `json:"data"`
	Timestamp time.Time `json:"ts,omitempty"`
}

// VaultResponse ...
type VaultResponse struct {
	Items   []*VaultItem `json:"items"`
	Version string       `json:"version"`
}
