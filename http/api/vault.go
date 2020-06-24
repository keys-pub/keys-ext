package api

import "time"

// VaultBox ...
type VaultBox struct {
	// Data encrypted.
	Data []byte `json:"data"`
	// Version of data from remote (untrusted).
	Version int64 `json:"v,omitempty"`
	// Timestamp when data was saved on the remote (untrusted).
	Timestamp time.Time `json:"ts,omitempty"`
}

// VaultResponse ...
type VaultResponse struct {
	Boxes   []*VaultBox `json:"datas"`
	Version string      `json:"version"`
}
