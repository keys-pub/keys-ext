package service

import (
	"context"
	"fmt"
	"time"

	"github.com/keys-pub/keys-ext/http/api"
)

// AdminSignURL ...
func (s *service) AdminSignURL(ctx context.Context, req *AdminSignURLRequest) (*AdminSignURLResponse, error) {
	key, err := s.parseIdentityForEdX25519Key(ctx, req.Signer)
	if err != nil {
		return nil, err
	}

	auth, err := api.NewAuth(req.Method, req.URL, time.Now(), key)
	if err != nil {
		return nil, err
	}

	curl := fmt.Sprintf("curl -X %s -H \"Authorization: %s\" %s", req.Method, auth.Header(), auth.URL.String())

	return &AdminSignURLResponse{
		Auth: auth.Header(),
		URL:  auth.URL.String(),
		CURL: curl,
	}, nil
}

func (s *service) AdminCheck(ctx context.Context, req *AdminCheckRequest) (*AdminCheckResponse, error) {
	admin, err := s.parseIdentityForEdX25519Key(ctx, req.Signer)
	if err != nil {
		return nil, err
	}

	if err := s.client.AdminCheck(ctx, req.Check, admin); err != nil {
		return nil, err
	}

	return &AdminCheckResponse{}, nil
}
