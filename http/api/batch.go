package api

import (
	"encoding/json"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/http"
)

// BatchRequests ...
type BatchRequests struct {
	Requests []*BatchRequest `json:"requests"`
}

// BatchRequest ...
type BatchRequest struct {
	ID      string            `json:"id"`
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

// NewBatchRequest returns batch request.
func NewBatchRequest(id string, method string, urs string, contentHash string, now time.Time, key *keys.EdX25519Key) (*BatchRequest, error) {
	auth, err := http.NewAuth(method, urs, contentHash, now, key)
	if err != nil {
		return nil, err
	}

	return &BatchRequest{
		ID:     id,
		Method: method,
		URL:    auth.URL.String(),
		Headers: map[string]string{
			"Authorization": auth.Header(),
		},
	}, nil
}

// BatchResponses ...
type BatchResponses struct {
	Responses []*BatchResponse `json:"responses"`
}

// BatchResponse ..
type BatchResponse struct {
	ID     string      `json:"id"`
	Status int         `json:"status"`
	Body   interface{} `json:"body"`
}

func (r *BatchResponse) Error() *Error {
	if r.Status >= 200 && r.Status < 300 {
		return nil
	}
	var out Error
	if err := r.As(&out); err != nil {
		return &Error{Status: r.Status}
	}
	return &out
}

// As unmarshals into value.
func (r *BatchResponse) As(v interface{}) error {
	b, err := json.Marshal(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
