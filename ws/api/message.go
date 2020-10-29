package api

import (
	"github.com/keys-pub/keys"
)

// Message ...
type Message struct {
	KID keys.ID `json:"kid"`
}
