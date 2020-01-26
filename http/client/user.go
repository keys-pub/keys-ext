package client

import (
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/keys-pub/keysd/http/api"
	"github.com/pkg/errors"
)

// UserSearch ...
func (c *Client) UserSearch(query string, limit int) (*api.UserSearchResponse, error) {
	params := url.Values{}
	params.Add("q", query)
	if limit > 0 {
		params.Add("limit", strconv.Itoa(limit))
	}

	e, err := c.get("/user/search", params, nil)
	if err != nil {
		return nil, err
	}
	if e == nil {
		return nil, errors.Errorf("/search not found")
	}

	var val api.UserSearchResponse
	if err := json.Unmarshal(e.Data, &val); err != nil {
		return nil, err
	}
	return &val, nil
}
