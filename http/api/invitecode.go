package api

import "github.com/keys-pub/keys"

// InviteCodeCreateResponse ...
type InviteCodeCreateResponse struct {
	Code string `json:"code"`
}

// InviteCodeResponse ...
type InviteCodeResponse struct {
	Sender    keys.ID `json:"sender"`
	Recipient keys.ID `json:"recipient"`
}
