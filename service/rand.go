package service

import (
	"context"
	"encoding/hex"

	"github.com/keys-pub/keys"
)

// Rand (RPC) ...
func (s *service) Rand(ctx context.Context, req *RandRequest) (*RandResponse, error) {
	b := keys.RandBytes(int(req.Length))
	str := ""

	if req.Encoding != Hex {
		enc, err := encodingFromRPC(req.Encoding)
		if err != nil {
			return nil, err
		}
		s, encErr := keys.Encode(b, enc)
		if encErr != nil {
			return nil, encErr
		}
		str = s
	} else {
		str = hex.EncodeToString(b)
	}

	return &RandResponse{
		Data: str,
	}, nil
}
