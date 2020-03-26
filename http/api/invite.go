package api

import "github.com/keys-pub/keys"

// CreateInviteResponse ...
type CreateInviteResponse struct {
	Code string `json:"code"`
}

// InviteResponse ...
type InviteResponse struct {
	Sender    keys.ID `json:"sender"`
	Recipient keys.ID `json:"recipient"`
}
