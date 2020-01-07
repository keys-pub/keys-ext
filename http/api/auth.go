package api

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/keys-pub/keys"
)

// Auth describes auth for a HTTP request.
type Auth struct {
	KID     keys.ID
	Method  string
	URL     *url.URL
	Sig     string
	Message string
}

// Header is header value.
func (a Auth) Header() string {
	return a.KID.String() + ":" + a.Sig
}

// NewAuth returns auth for an HTTP request.
// The url shouldn't have ? or &.
func NewAuth(method string, urs string, tm time.Time, key *keys.SignKey) (*Auth, error) {
	return newAuth(method, urs, tm, keys.Rand32(), key)
}

func newAuth(method string, urs string, tm time.Time, nonce *[32]byte, key *keys.SignKey) (*Auth, error) {
	ur, err := url.Parse(urs)
	if err != nil {
		return nil, err
	}
	q := ur.Query()
	ns := keys.MustEncode(nonce[:], keys.Base62)
	q.Set("nonce", ns)
	ts := keys.TimeToMillis(tm)
	q.Set("ts", fmt.Sprintf("%d", ts))
	ur.RawQuery = q.Encode()

	msg := method + "," + ur.String()
	sb := key.SignDetached([]byte(msg))
	sig := keys.MustEncode(sb, keys.Base62)
	return &Auth{KID: key.ID(), Method: method, URL: ur, Sig: sig, Message: msg}, nil
}

// NewRequest returns new authorized/signed HTTP request.
func NewRequest(method string, urs string, body io.Reader, tm time.Time, key *keys.SignKey) (*http.Request, error) {
	return newRequest(method, urs, body, tm, keys.Rand32(), key)
}

func newRequest(method string, urs string, body io.Reader, tm time.Time, nonce *[32]byte, key *keys.SignKey) (*http.Request, error) {
	auth, err := newAuth(method, urs, tm, nonce, key)
	if err != nil {
		return nil, err
	}
	logger.Infof("Auth for %s", auth.Message)
	req, err := http.NewRequest(method, auth.URL.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", auth.Header())
	return req, nil
}
