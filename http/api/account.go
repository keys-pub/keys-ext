package api

import (
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/vault/auth/api"
)

// Account ...
type Account struct {
	Email string  `json:"email"`
	KID   keys.ID `json:"kid"`

	VerifyEmailCode   string    `json:"verifyEmailCode"`
	VerifyEmailCodeAt time.Time `json:"verifyEmailCodeAt"`
	VerifiedEmail     bool      `json:"verifiedEmail"`
	VerifiedEmailAt   time.Time `json:"verifiedEmailAt"`
}

// SendEmailVerificationResponse ...
type SendEmailVerificationResponse struct {
	Email string  `json:"email"`
	KID   keys.ID `json:"kid"`
}

// AccountCreateRequest ...
type AccountCreateRequest struct {
	Email string `json:"email"`
}

// AccountCreateResponse ...
type AccountCreateResponse struct {
	Email string  `json:"email"`
	KID   keys.ID `json:"kid"`
}

// AccountResponse ...
type AccountResponse struct {
	Email         string  `json:"email"`
	KID           keys.ID `json:"kid"`
	VerifiedEmail bool    `json:"verifiedEmail"`
}

// AccountVerifyEmailRequest ...
type AccountVerifyEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

// AccountVault ...
type AccountVault struct {
	AID   keys.ID `json:"aid"`
	VID   keys.ID `json:"vid"`
	Token string  `json:"token"`
	Usage int64   `json:"usage"`
}

// AccountVaultsResponse ...
type AccountVaultsResponse struct {
	Vaults []*AccountVault `json:"vaults"`
}

type Auth = api.Auth

type AccountAuth struct {
	AID  keys.ID `json:"aid"`
	Auth *Auth   `json:"auth"`
}

type AccountAuthsResponse struct {
	Auths []*Auth `json:"auths"`
}
