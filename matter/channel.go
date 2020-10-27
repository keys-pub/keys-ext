package matter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

// Channel constants.
const (
	ChannelOpen    = "O"
	ChannelPrivate = "P"
	ChannelDirect  = "D"
	ChannelGroup   = "G"
	// CHANNEL_GROUP_MAX_USERS        = 8
	// CHANNEL_GROUP_MIN_USERS        = 3
	DefaultChannel = "town-square"
	// CHANNEL_DISPLAY_NAME_MAX_RUNES = 64
	// CHANNEL_NAME_MIN_LENGTH        = 2
	// CHANNEL_NAME_MAX_LENGTH        = 64
	// CHANNEL_HEADER_MAX_RUNES       = 1024
	// CHANNEL_PURPOSE_MAX_RUNES      = 250
	// CHANNEL_CACHE_SIZE             = 25000

	// CHANNEL_SORT_BY_USERNAME = "username"
	// CHANNEL_SORT_BY_STATUS   = "status"
)

// CreateChannel creates a channel.
func (c *Client) CreateChannel(ctx context.Context, name string, teamID string) (*Channel, error) {
	channel := &Channel{
		TeamID: teamID,
		Name:   name,
	}
	b, err := json.Marshal(channel)
	if err != nil {
		return nil, err
	}
	resp, err := c.Post(ctx, "/api/v4/channels", nil, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	var out Channel
	if err := unmarshal(resp, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Channel finds channel by ID.
func (c *Client) Channel(ctx context.Context, teamID string, id string) (*Channel, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/v4/teams/%s/channels/%s", teamID, id), nil)
	if err != nil {
		return nil, err
	}
	var channel Channel
	if err := unmarshal(resp, &channel); err != nil {
		return nil, err
	}
	return &channel, nil
}

// ChannelByName finds channel by name.
func (c *Client) ChannelByName(ctx context.Context, teamID string, name string) (*Channel, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/v4/teams/%s/channels/name/%s", teamID, name), nil)
	if err != nil {
		return nil, err
	}
	var channel Channel
	if err := unmarshal(resp, &channel); err != nil {
		return nil, err
	}
	return &channel, nil
}

// AddUserToChannel adds a user to a channel.
func (c *Client) AddUserToChannel(ctx context.Context, userID string, channelID string) error {
	params := map[string]string{}
	params["user_id"] = userID
	params["channel_id"] = channelID
	b, err := json.Marshal(params)
	if err != nil {
		return err
	}
	_, err = c.Post(ctx, fmt.Sprintf("/api/v4/channels/%s/members", channelID), nil, bytes.NewReader(b))
	if err != nil {
		return err
	}
	return nil
}

// ChannelsForUser list channels for logged in user.
// If userID is "", logged in user (me) is used.
func (c *Client) ChannelsForUser(ctx context.Context, userID string, teamID string) ([]*Channel, error) {
	if userID == "" {
		userID = "me"
	}
	resp, err := c.Get(ctx, fmt.Sprintf("/api/v4/users/%s/teams/%s/channels", userID, teamID), nil)
	if err != nil {
		return nil, err
	}
	var channels []*Channel
	if err := unmarshal(resp, &channels); err != nil {
		return nil, err
	}
	return channels, nil
}
