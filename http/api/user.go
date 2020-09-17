package api

import (
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/user"
	"github.com/keys-pub/keys/users"
)

// UserFromSearchResult returns User from user.SearchResult.
func UserFromSearchResult(sr *users.SearchResult) *User {
	if sr == nil {
		return nil
	}
	user := UserFromResult(sr.Result)
	if user != nil {
		user.MatchField = sr.Field
	}
	return user
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

// UserResponse ...
type UserResponse struct {
	User *User `json:"user"`
}

// UserSearchResponse ...
type UserSearchResponse struct {
	Users []*User `json:"users"`
}
