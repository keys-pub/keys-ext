package client

import (
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/api"
	"github.com/pkg/errors"
)

// Search ...
func (c *Client) Search(query string, limit int) (*api.SearchResponse, error) {
	params := url.Values{}
	params.Add("q", query)
	if limit > 0 {
		params.Add("limit", strconv.Itoa(limit))
	}

	e, err := c.get(keys.Path("search"), params, nil)
	if err != nil {
		return nil, err
	}
	if e == nil {
		return nil, errors.Errorf("/search not found")
	}

	var val api.SearchResponse
	if err := json.Unmarshal(e.Data, &val); err != nil {
		return nil, err
	}
	return &val, nil
}
