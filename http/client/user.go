package client

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/api"
	"github.com/pkg/errors"
)

// UserSearch ...
func (c *Client) UserSearch(ctx context.Context, query string, limit int) (*api.UserSearchResponse, error) {
	params := url.Values{}
	params.Add("q", query)
	if limit > 0 {
		params.Add("limit", strconv.Itoa(limit))
	}

	doc, err := c.getDocument(ctx, "/user/search", params, nil)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, errors.Errorf("/user/search not found")
	}

	var val api.UserSearchResponse
	if err := json.Unmarshal(doc.Data, &val); err != nil {
		return nil, err
	}
	return &val, nil
}

// User ...
func (c *Client) User(ctx context.Context, kid keys.ID) (*api.UserResponse, error) {
	params := url.Values{}
	doc, err := c.getDocument(ctx, "/user/"+kid.String(), params, nil)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}

	var val api.UserResponse
	if err := json.Unmarshal(doc.Data, &val); err != nil {
		return nil, err
	}
	return &val, nil
}
