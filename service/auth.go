package service

import (
	"context"
	"sync"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keys/keyring"
	"github.com/keys-pub/keysd/auth/fido2"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TODO: Some clients log grpc requests which for AuthSetup and AuthUnlock include a password.
//       We need to ensure client logging doesn't do this in the future accidentally.

type auth struct {
	sync.Mutex
	cfg       *Config
	keyring   *keyring.Keyring
	tokens    map[string]string
	whitelist *ds.StringSet

	auths fido2.AuthServer
}

func newAuth(cfg *Config, st keyring.Store) (*auth, error) {
	// We don't need auth for the following methods.
	whitelist := ds.NewStringSet(
		"/service.Keys/AuthSetup",
		"/service.Keys/AuthUnlock",
		"/service.Keys/AuthLock",
		"/service.Keys/RuntimeStatus",
		"/service.Keys/Rand",
		"/service.Keys/RandPassword",
	)

	service := cfg.KeyringService(st.Name())
	kr, err := keyring.New(service, st)
	if err != nil {
		return nil, err
	}

	return &auth{
		cfg:       cfg,
		keyring:   kr,
		tokens:    map[string]string{},
		whitelist: whitelist,
	}, nil
}

func (a *auth) lock() error {
	// TODO: Lock after running for a certain amount of time (maybe a few hours?)
	logger.Infof("Locking")
	if err := a.keyring.Lock(); err != nil {
		return err
	}
	a.tokens = map[string]string{}
	return nil
}

func (a *auth) setup(ctx context.Context, secret string, typ AuthType) error {
	logger.Infof("Setup (%s)", typ)
	var err error
	switch typ {
	case PasswordAuth:
		err = a.setupPassword(secret)
	case FIDO2HMACSecretAuth:
		_, err = a.setupHMACSecret(ctx, secret)
	default:
		return errors.Errorf("unsupported auth type")
	}
	if err != nil {
		return authErr(err, typ, "failed to setup")
	}
	return nil
}

func (a *auth) unlock(ctx context.Context, secret string, typ AuthType, client string) (string, error) {
	logger.Infof("Unlock (%s)", typ)
	var err error
	switch typ {
	case PasswordAuth:
		err = a.unlockPassword(secret)
	case FIDO2HMACSecretAuth:
		err = a.unlockHMACSecret(ctx, secret)
	default:
		return "", errors.Errorf("unsupported auth type")
	}
	if err != nil {
		return "", authErr(err, typ, "failed to unlock")
	}
	logger.Infof("Unlocked (%s)", typ)
	token := a.registerToken(client)
	return token, nil
}

func (a *auth) provision(ctx context.Context, secret string, typ AuthType, setup bool) (string, error) {
	logger.Infof("Provision (%s)", typ)
	switch typ {
	case PasswordAuth:
		return a.provisionPassword(ctx, secret)
	case FIDO2HMACSecretAuth:
		if setup {
			return a.setupHMACSecret(ctx, secret)
		}
		return a.provisionHMACSecret(ctx, secret)
	default:
		return "", errors.Errorf("unknown auth type")
	}
}

func (a *auth) registerToken(client string) string {
	token := generateToken()
	a.tokens[client] = token
	return token
}

func authErr(err error, typ AuthType, wrap string) error {
	if errors.Cause(err) == keyring.ErrInvalidAuth {
		switch typ {
		case PasswordAuth:
			return status.Error(codes.Unauthenticated, "invalid password")
		default:
			return status.Error(codes.Unauthenticated, "invalid auth")
		}

	}
	return errors.Wrapf(err, wrap)
}

func generateToken() string {
	return encoding.MustEncode(keys.Rand32()[:], encoding.Base62)
}

func (a *auth) streamInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := a.authorize(stream.Context(), info.FullMethod); err != nil {
		return err
	}
	return handler(srv, stream)
}

func (a *auth) unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if err := a.authorize(ctx, info.FullMethod); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

func (a *auth) authorize(ctx context.Context, method string) error {
	// No authorization needed for whitelisted methods.
	if a.whitelist.Contains(method) {
		logger.Infof("Authorization is not required for %s", method)
		return nil
	}

	logger.Infof("Authorize %s", method)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if len(md["authorization"]) == 0 {
			logger.Warningf("Auth token missing from request")
			return status.Error(codes.Unauthenticated, "authorization missing")
		}
		token := md["authorization"][0]
		for _, t := range a.tokens {
			if t == token {
				return nil
			}
		}

		logger.Infof("Invalid auth token")
		return status.Error(codes.Unauthenticated, "invalid token")
	}
	return status.Error(codes.Unauthenticated, "no authorization in context")
}

type clientAuth struct {
	token string
}

func newClientAuth(token string) clientAuth {
	return clientAuth{token: token}
}

func (a clientAuth) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	if a.token == "" {
		return nil, nil
	}
	return map[string]string{
		"authorization": a.token,
	}, nil
}

func (a clientAuth) RequireTransportSecurity() bool {
	return true
}
