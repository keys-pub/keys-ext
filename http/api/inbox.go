package api

// InboxChannelsResponse ...
type InboxChannelsResponse struct {
	Channels []*Channel `json:"channels"`
}

// InboxChannelInviteResponse ...
type InboxChannelInviteResponse struct {
	Invite *ChannelInvite `json:"invite"`
}

// InboxChannelInvitesResponse ...
type InboxChannelInvitesResponse struct {
	Invites []*ChannelInvite `json:"invites"`
}
