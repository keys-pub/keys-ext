package server_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/keys-pub/keys/http"
	"github.com/pkg/errors"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
)

type firebaseAuth struct {
	client *auth.Client
	apiKey string
}

func (f firebaseAuth) CreateEmailUser(ctx context.Context, email string, password string) (string, error) {
	params := (&auth.UserToCreate{}).
		Email(email).
		Password(password)
	u, err := f.client.CreateUser(ctx, params)
	if err != nil {
		return "", err
	}
	return u.UID, nil
}

func (f firebaseAuth) SendEmailVerification(ctx context.Context, email string, password string) error {
	// Sign in to get token
	signInReq, err := http.NewRequestWithContext(ctx,
		"POST",
		"https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key="+f.apiKey,
		bytes.NewReader([]byte(`{"email":"`+email+`","password":"`+password+`","returnSecureToken":true}`)))
	if err != nil {
		return err
	}
	var out struct {
		IDToken string `json:"idToken"`
		Email   string `json:"email"`
	}
	signInReq.Header.Set("Content-Type", "application/json")
	if err := http.JSON(signInReq, &out); err != nil {
		return errors.Wrapf(err, "failed to sign in with identitytoolkit")
	}
	token := out.IDToken

	// Send email request
	emailReq, err := http.NewRequestWithContext(ctx,
		"POST",
		"https://identitytoolkit.googleapis.com/v1/accounts:sendOobCode?key="+f.apiKey,
		bytes.NewReader([]byte(`{"requestType":"VERIFY_EMAIL","idToken":"`+token+`"}`)))
	if err != nil {
		return err
	}
	emailReq.Header.Set("Content-Type", "application/json")
	if err := http.JSON(emailReq, &out); err != nil {
		return errors.Wrapf(err, "failed to send email with identitytoolkit")
	}
	return nil
}

func TestFirebaseUser(t *testing.T) {
	opt := option.WithCredentialsFile("credentials-firebase.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	require.NoError(t, err)

	client, err := app.Auth(context.Background())
	require.NoError(t, err)
	apiKey := "AIzaSyBYl4tVgYuCCK5EYsKrMj49OMUC8JU8jvs"

	fi := &firebaseAuth{client, apiKey}

	_, err = fi.CreateEmailUser(context.TODO(), "test2@rel.me", "soiihiwqjeilehetjhgsdfaa234553")
	require.NoError(t, err)

	err = fi.SendEmailVerification(context.TODO(), "test2@rel.me", "soiihiwqjeilehetjhgsdfaa234553")
	require.NoError(t, err)
}

// func TestFirebaseServerAuth(t *testing.T) {
// 	env := newEnv(t)
// 	srv := newTestServer(t, env)
// 	clock := env.clock
// 	// srv.Server.SetFirebaseAuth(auth)

// 	alice := keys.NewEdX25519KeyFromSeed(testSeed(0x01))

// 	b, err := json.Marshal(&server.AccountCreateRequest{
// 		// KID:   alice.ID(),
// 		// Email: "test@rel.me",
// 	})
// 	require.NoError(t, err)

// 	// PUT /accounts/:cid
// 	req, err := http.NewAuthRequest("PUT", dstore.Path("accounts", alice.ID()), bytes.NewReader(b), http.ContentHash(b), clock.Now(), alice)
// 	require.NoError(t, err)
// 	code, _, body := srv.Serve(req)
// 	var create server.AccountCreateResponse
// 	err = json.Unmarshal([]byte(body), &create)
// 	require.NoError(t, err)
// 	require.Equal(t, http.StatusOK, code)
// 	// require.Equal(t, alice.ID(), create.KID)
// }
