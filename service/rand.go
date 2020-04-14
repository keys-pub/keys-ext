package service

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/encoding"
)

// Rand (RPC) ...
func (s *service) Rand(ctx context.Context, req *RandRequest) (*RandResponse, error) {
	b := keys.RandBytes(int(req.NumBytes))

	enc, err := encodingFromRPC(req.Encoding)
	if err != nil {
		return nil, err
	}
	out, err := encoding.Encode(b, enc)
	if err != nil {
		return nil, err
	}

	return &RandResponse{
		Data: out,
	}, nil
}

func (s *service) RandPassword(ctx context.Context, req *RandPasswordRequest) (*RandPasswordResponse, error) {
	password := keys.RandPassword(int(req.Length))
	return &RandPasswordResponse{
		Password: password,
	}, nil
}
