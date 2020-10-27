package matter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

// Post constants.
const (
	PostSystemMessagePrefix = "system_"
	PostDefault             = ""
	// POST_SLACK_ATTACHMENT       = "slack_attachment"
	// POST_SYSTEM_GENERIC         = "system_generic"
	// POST_JOIN_LEAVE             = "system_join_leave" // Deprecated, use POST_JOIN_CHANNEL or POST_LEAVE_CHANNEL instead
	// POST_JOIN_CHANNEL           = "system_join_channel"
	// POST_GUEST_JOIN_CHANNEL     = "system_guest_join_channel"
	// POST_LEAVE_CHANNEL          = "system_leave_channel"
	// POST_JOIN_TEAM              = "system_join_team"
	// POST_LEAVE_TEAM             = "system_leave_team"
	// POST_AUTO_RESPONDER         = "system_auto_responder"
	// POST_ADD_REMOVE             = "system_add_remove" // Deprecated, use POST_ADD_TO_CHANNEL or POST_REMOVE_FROM_CHANNEL instead
	// POST_ADD_TO_CHANNEL         = "system_add_to_channel"
	// POST_ADD_GUEST_TO_CHANNEL   = "system_add_guest_to_chan"
	// POST_REMOVE_FROM_CHANNEL    = "system_remove_from_channel"
	// POST_MOVE_CHANNEL           = "system_move_channel"
	// POST_ADD_TO_TEAM            = "system_add_to_team"
	// POST_REMOVE_FROM_TEAM       = "system_remove_from_team"
	// POST_HEADER_CHANGE          = "system_header_change"
	// POST_DISPLAYNAME_CHANGE     = "system_displayname_change"
	// POST_CONVERT_CHANNEL        = "system_convert_channel"
	// POST_PURPOSE_CHANGE         = "system_purpose_change"
	// POST_CHANNEL_DELETED        = "system_channel_deleted"
	// POST_CHANNEL_RESTORED       = "system_channel_restored"
	// POST_EPHEMERAL              = "system_ephemeral"
	// POST_CHANGE_CHANNEL_PRIVACY = "system_change_chan_privacy"
	// POST_ADD_BOT_TEAMS_CHANNELS = "add_bot_teams_channels"
	// POST_FILEIDS_MAX_RUNES      = 150
	// POST_FILENAMES_MAX_RUNES    = 4000
	// POST_HASHTAGS_MAX_RUNES     = 1000
	// POST_MESSAGE_MAX_RUNES_V1   = 4000
	// POST_MESSAGE_MAX_BYTES_V2   = 65535                         // Maximum size of a TEXT column in MySQL
	// POST_MESSAGE_MAX_RUNES_V2   = POST_MESSAGE_MAX_BYTES_V2 / 4 // Assume a worst-case representation
	// POST_PROPS_MAX_RUNES        = 8000
	// POST_PROPS_MAX_USER_RUNES   = POST_PROPS_MAX_RUNES - 400 // Leave some room for system / pre-save modifications
	// POST_CUSTOM_TYPE_PREFIX     = "custom_"
	// POST_ME                     = "me"
	// PROPS_ADD_CHANNEL_MEMBER    = "add_channel_member"

	// POST_PROPS_ADDED_USER_ID       = "addedUserId"
	// POST_PROPS_DELETE_BY           = "deleteBy"
	// POST_PROPS_OVERRIDE_ICON_URL   = "override_icon_url"
	// POST_PROPS_OVERRIDE_ICON_EMOJI = "override_icon_emoji"

	// POST_PROPS_MENTION_HIGHLIGHT_DISABLED = "mentionHighlightDisabled"
	// POST_PROPS_GROUP_HIGHLIGHT_DISABLED   = "disable_group_highlight"
	// POST_SYSTEM_WARN_METRIC_STATUS        = "warn_metric_status"
)

// PostsForChannel returns posts for channel.
func (c *Client) PostsForChannel(ctx context.Context, channelID string) (*PostList, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/v4/channels/%s/posts", channelID), nil)
	if err != nil {
		return nil, err
	}
	var posts *PostList
	if err := unmarshal(resp, &posts); err != nil {
		return nil, err
	}
	return posts, nil

}

// CreatePost creates a post.
func (c *Client) CreatePost(ctx context.Context, channelID string, message string) (*Post, error) {
	post := &Post{
		ChannelID: channelID,
		Message:   message,
	}
	b, err := json.Marshal(post)
	if err != nil {
		return nil, err
	}
	resp, err := c.Post(ctx, "/api/v4/posts", nil, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	var out Post
	if err := unmarshal(resp, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
