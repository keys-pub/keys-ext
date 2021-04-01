package api

import (
	"time"

	"github.com/keys-pub/keys"
)

// Account ...
type Account struct {
	KID   keys.ID `json:"kid"`
	Email string  `json:"email"`

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

type AccountAuth struct {
	ID   string `json:"id"`
	Data []byte `json:"data"` // Encrypted auth data
}

type AccountAuthsResponse struct {
	Auths []*AccountAuth `json:"auths"`
}
