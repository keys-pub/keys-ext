package api

import (
	"time"
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
