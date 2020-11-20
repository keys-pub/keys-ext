package api

// UserChannelsResponse ...
type UserChannelsResponse struct {
	Channels []*Channel `json:"channels"`
}

// UserChannelInviteResponse ...
type UserChannelInviteResponse struct {
	Invite *ChannelInvite `json:"invite"`
}

// UserChannelInvitesResponse ...
type UserChannelInvitesResponse struct {
	Invites []*ChannelInvite `json:"invites"`
}
