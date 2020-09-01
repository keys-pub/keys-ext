package client

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/pkg/errors"
)

// UserSearch ...
func (c *Client) UserSearch(ctx context.Context, query string, limit int) (*api.UserSearchResponse, error) {
	params := url.Values{}
	params.Add("q", query)
	if limit > 0 {
		params.Add("limit", strconv.Itoa(limit))
	}

	doc, err := c.getDocument(ctx, "/users/search", params, nil)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, errors.Errorf("/users/search not found")
	}

	// TODO: Support paging
	var resp api.UserSearchResponse
	if err := json.Unmarshal(doc.Data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Users ...
func (c *Client) Users(ctx context.Context, kid keys.ID) (*api.UsersResponse, error) {
	params := url.Values{}
	doc, err := c.getDocument(ctx, "/users/"+kid.String(), params, nil)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}

	var resp api.UsersResponse
	if err := json.Unmarshal(doc.Data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
