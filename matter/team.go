package matter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

// Team constants.
const (
	TeamOpen   string = "O"
	TeamInvite string = "I"
)

// CreateTeam creates a team.
func (c *Client) CreateTeam(ctx context.Context, name string, displayName string, typ string) (*Team, error) {
	team := &Team{
		Name:        name,
		DisplayName: displayName,
		Type:        typ,
	}
	b, err := json.Marshal(team)
	if err != nil {
		return nil, err
	}
	resp, err := c.Post(ctx, "/api/v4/teams", nil, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	var out Team
	if err := unmarshal(resp, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TeamsForUser lists teams for user.
// If userID is "", logged in user (me) is used.
func (c *Client) TeamsForUser(ctx context.Context, userID string) ([]*Team, error) {
	if userID == "" {
		userID = "me"
	}
	resp, err := c.Get(ctx, fmt.Sprintf("/api/v4/users/%s/teams", userID), nil)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	var teams []*Team
	if err := unmarshal(resp, &teams); err != nil {
		return nil, err
	}
	return teams, nil
}

// Teams ...
func (c *Client) Teams(ctx context.Context) ([]*Team, error) {
	resp, err := c.Get(ctx, "/api/v4/teams", nil)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	var teams []*Team
	if err := unmarshal(resp, &teams); err != nil {
		return nil, err
	}
	return teams, nil
}

// TeamByName finds team.
func (c *Client) TeamByName(ctx context.Context, name string) (*Team, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/v4/teams/name/%s", name), nil)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	var team Team
	if err := unmarshal(resp, &team); err != nil {
		return nil, err
	}
	return &team, nil
}

// // TeamExists checks if team exists.
// func (c *Client) TeamExists(ctx context.Context, name string) (bool, error) {
// 	resp, err := c.Get(ctx, fmt.Sprintf("/api/v4/teams/name/%s/exists", name), nil)
// 	if err != nil {
// 		return false, err
// 	}
// 	var m map[string]bool
// 	if err := unmarshal(resp, &m); err != nil {
// 		return false, err
// 	}
// 	exists, _ := m["exists"]
// 	return exists, nil
// }

// AddUserToTeam adds a user to a team.
func (c *Client) AddUserToTeam(ctx context.Context, userID string, teamID string) error {
	params := map[string]string{}
	params["user_id"] = userID
	params["team_id"] = teamID
	b, err := json.Marshal(params)
	if err != nil {
		return err
	}
	_, err = c.Post(ctx, fmt.Sprintf("/api/v4/teams/%s/members", teamID), nil, bytes.NewReader(b))
	if err != nil {
		return err
	}
	return nil
}
