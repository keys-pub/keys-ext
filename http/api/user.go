package api

import (
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/user"
)

// UserFromSearchResult returns User from user.SearchResult.
func UserFromSearchResult(sr *user.SearchResult) *User {
	if sr == nil {
		return nil
	}
	user := UserFromResult(sr.Result)
	if user != nil {
		user.MatchField = sr.Field
	}
	return user
}

// UsersFromResults returns []*User from []*user.Result.
func UsersFromResults(results []*user.Result) []*User {
	users := make([]*User, 0, len(results))
	for _, r := range results {
		users = append(users, UserFromResult(r))
	}
	return users
}

// UserFromResult returns User from user.Result.
func UserFromResult(result *user.Result) *User {
	if result == nil {
		return nil
	}
	return &User{
		ID:         result.User.Name + "@" + result.User.Service,
		KID:        result.User.KID,
		Seq:        result.User.Seq,
		Service:    result.User.Service,
		Name:       result.User.Name,
		URL:        result.User.URL,
		Status:     result.Status,
		VerifiedAt: result.VerifiedAt,
		Timestamp:  result.Timestamp,
		Err:        result.Err,
	}
}

// User ...
type User struct {
	ID         string      `json:"id,omitempty"`
	Name       string      `json:"name,omitempty"`
	KID        keys.ID     `json:"kid,omitempty"`
	Seq        int         `json:"seq,omitempty"`
	Service    string      `json:"service,omitempty"`
	URL        string      `json:"url,omitempty"`
	Status     user.Status `json:"status,omitempty"`
	VerifiedAt int64       `json:"verifiedAt,omitempty"`
	Timestamp  int64       `json:"ts,omitempty"`
	MatchField string      `json:"mf,omitempty"`
	Err        string      `json:"err,omitempty"`
}

// UsersResponse ...
type UsersResponse struct {
	Users []*User `json:"users"`
}

// UserSearchResponse ...
type UserSearchResponse struct {
	Users []*User `json:"users"`
}
