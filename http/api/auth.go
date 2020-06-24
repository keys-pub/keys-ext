package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
)

// TODO: Headers are not included in auth.

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
func NewAuth(method string, urs string, tm time.Time, key *keys.EdX25519Key) (*Auth, error) {
	return newAuth(method, urs, tm, keys.Rand32(), key)
}

func newAuth(method string, urs string, tm time.Time, nonce *[32]byte, key *keys.EdX25519Key) (*Auth, error) {
	ur, err := url.Parse(urs)
	if err != nil {
		return nil, err
	}
	q := ur.Query()
	ns := encoding.MustEncode(nonce[:], encoding.Base62)
	q.Set("nonce", ns)
	ts := tsutil.Millis(tm)
	q.Set("ts", fmt.Sprintf("%d", ts))
	ur.RawQuery = q.Encode()

	msg := method + "," + ur.String()
	logger.Debugf("Signing %s", msg)
	sb := key.SignDetached([]byte(msg))
	sig := encoding.MustEncode(sb, encoding.Base62)
	return &Auth{KID: key.ID(), Method: method, URL: ur, Sig: sig, Message: msg}, nil
}

// NewRequest returns new authorized/signed HTTP request.
func NewRequest(method string, urs string, body io.Reader, tm time.Time, key *keys.EdX25519Key) (*http.Request, error) {
	return newRequest(context.TODO(), method, urs, body, tm, keys.Rand32(), key)
}

// NewRequestWithContext returns new authorized/signed HTTP request with context.
func NewRequestWithContext(ctx context.Context, method string, urs string, body io.Reader, tm time.Time, key *keys.EdX25519Key) (*http.Request, error) {
	return newRequest(ctx, method, urs, body, tm, keys.Rand32(), key)
}

func newRequest(ctx context.Context, method string, urs string, body io.Reader, tm time.Time, nonce *[32]byte, key *keys.EdX25519Key) (*http.Request, error) {
	auth, err := newAuth(method, urs, tm, nonce, key)
	if err != nil {
		return nil, err
	}
	logger.Infof("Auth for %s", auth.Message)
	req, err := http.NewRequestWithContext(ctx, method, auth.URL.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", auth.Header())
	return req, nil
}

// AuthResult is the authorized result.
type AuthResult struct {
	KID       keys.ID
	Method    string
	URL       *url.URL
	Nonce     string
	Timestamp time.Time
}

// MemCache ...
type MemCache interface {
	// Get returns value at key.
	Get(ctx context.Context, k string) (string, error)
	// Put puts a value at key.
	Set(ctx context.Context, k string, v string) error
	// Expire key.
	Expire(ctx context.Context, k string, dt time.Duration) error
}

// CheckAuthorization checks auth header.
func CheckAuthorization(ctx context.Context, method string, urs string, auth string, mc MemCache, now time.Time) (*AuthResult, error) {
	fields := strings.Split(auth, ":")
	if len(fields) != 2 {
		return nil, errors.Errorf("too many fields")
	}
	skid := fields[0]
	sig := fields[1]

	kid, err := keys.ParseID(skid)
	if err != nil {
		return nil, err
	}

	spk, err := keys.StatementPublicKeyFromID(kid)
	if err != nil {
		return nil, errors.Errorf("not a valid sign public key")
	}

	sigBytes, sigerr := encoding.Decode(sig, encoding.Base62)
	if sigerr != nil {
		return nil, sigerr
	}

	url, err := url.Parse(urs)
	if err != nil {
		return nil, err
	}

	msg := method + "," + url.String()
	logger.Infof("Checking auth for %s %s", msg, auth)
	if err := spk.VerifyDetached(sigBytes, []byte(msg)); err != nil {
		return nil, err
	}

	nonce := url.Query().Get("nonce")
	if nonce == "" {
		return nil, errors.Errorf("nonce is missing")
	}
	nb, err := encoding.Decode(nonce, encoding.Base62)
	if err != nil {
		return nil, err
	}
	if len(nb) != 32 {
		return nil, errors.Errorf("nonce is invalid length")
	}

	// Check the nonce
	nonce = fmt.Sprintf("auth-%s", nonce)

	val, err := mc.Get(ctx, nonce)
	if err != nil {
		return nil, err
	}
	if val != "" {
		return nil, errors.Errorf("nonce collision")
	}
	if err := mc.Set(ctx, nonce, "1"); err != nil {
		return nil, err
	}
	if err := mc.Expire(ctx, nonce, time.Hour); err != nil {
		return nil, err
	}

	// Check timestamp
	ts := url.Query().Get("ts")
	if ts == "" {
		return nil, errors.Errorf("timestamp (ts) is missing")
	}
	i, err := strconv.Atoi(ts)
	if err != nil {
		return nil, err
	}
	tm := tsutil.ParseMillis(int64(i))
	td := now.Sub(tm)
	if td < 0 {
		td = td * -1
	}
	if td > 30*time.Minute {
		return nil, errors.Errorf("timestamp is invalid, diff %s", td)
	}

	logger.Infof("Auth OK %s", kid)
	return &AuthResult{
		KID:       kid,
		Method:    method,
		URL:       url,
		Nonce:     nonce,
		Timestamp: tm,
	}, nil
}

// GenerateNonce creates a nonce.
func GenerateNonce() string {
	return encoding.MustEncode(keys.RandBytes(32), encoding.Base62)
}
