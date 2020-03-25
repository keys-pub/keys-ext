package api

import "github.com/keys-pub/keys"

// EphemResponse ...
type EphemResponse struct {
	Code string `json:"code,omitempty"`
}

// InviteResponse ...
type InviteResponse struct {
	Sender    keys.ID `json:"sender"`
	Recipient keys.ID `json:"recipient"`
}
