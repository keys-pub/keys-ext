package service

import (
	"bytes"
	"context"
	"io"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/saltpack"
	"github.com/pkg/errors"
)

// Sign (RPC) ...
func (s *service) Sign(ctx context.Context, req *SignRequest) (*SignResponse, error) {
	key, err := s.parseSigner(req.Signer, true)
	if err != nil {
		return nil, err
	}

	signed, err := saltpack.Sign(req.Data, req.Armored, key)
	if err != nil {
		return nil, err
	}

	return &SignResponse{
		Data: signed,
		KID:  key.ID().String(),
	}, nil
}

// SignFile (RPC) ...
func (s *service) SignFile(srv Keys_SignFileServer) error {
	req, err := srv.Recv()
	if err != nil {
		return err
	}
	in := req.In
	if in == "" {
		return errors.Errorf("in not specified")
	}
	out := req.Out
	if out == "" {
		if req.Detached {
			out = in + ".sig"
		} else {
			out = in + ".signed"
		}
	}

	key, err := s.parseSigner(req.Signer, true)
	if err != nil {
		return err
	}

	if err := saltpack.SignFile(in, out, key, req.Armored, req.Detached); err != nil {
		return err
	}

	if err := srv.Send(&SignFileOutput{
		KID: key.ID().String(),
		Out: out,
	}); err != nil {
		return err
	}

	return nil
}

// SignStream (RPC) ...
func (s *service) SignStream(srv Keys_SignStreamServer) error {
	init := false

	var stream io.WriteCloser
	var buf bytes.Buffer

	ctx := srv.Context()
	var kid keys.ID
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req, recvErr := srv.Recv()
		if recvErr == io.EOF {
			break
		}
		if recvErr != nil {
			return recvErr
		}

		if !init {
			init = true
			if stream != nil {
				return errors.Errorf("stream already initialized")
			}
			key, err := s.parseSigner(req.Signer, true)
			if err != nil {
				return err
			}
			logger.Debugf("Sign armored=%t detached=%t", req.Armored, req.Detached)
			s, err := saltpack.NewSignStream(&buf, req.Armored, req.Detached, key)
			if err != nil {
				return err
			}
			stream = s
			kid = key.ID()

		} else {
			// Make sure request only sends data after init
			if req.Signer != "" || req.Armored || req.Detached {
				return errors.Errorf("after stream is initalized, only data should be sent")
			}
		}

		if len(req.Data) > 0 {
			n, writeErr := stream.Write(req.Data)
			if writeErr != nil {
				return writeErr
			}
			if n != len(req.Data) {
				return errors.Errorf("failed to write completely (%d != %d)", n, len(req.Data))
			}

			if buf.Len() > 0 {
				out := buf.Bytes()
				if err := srv.Send(&SignOutput{
					Data: out,
					KID:  kid.String(),
				}); err != nil {
					return err
				}
				buf.Reset()
			}
		}
	}
	// Close stream and flush buffer
	if err := stream.Close(); err != nil {
		return err
	}
	if buf.Len() > 0 {
		out := buf.Bytes()
		if err := srv.Send(&SignOutput{
			Data: out,
			KID:  kid.String(),
		}); err != nil {
			return err
		}
		buf.Reset()
	}
	return nil
}
