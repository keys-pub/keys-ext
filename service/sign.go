package service

import (
	"bytes"
	"context"
	"io"

	"github.com/keys-pub/keys/saltpack"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Sign (RPC) ...
func (s *service) Sign(ctx context.Context, req *SignRequest) (*SignResponse, error) {
	key, err := s.parseKeyOrCurrent(req.KID)
	if err != nil {
		return nil, err
	}

	sp := saltpack.NewSaltpack(s.ks)
	sp.SetArmored(req.Armored)
	signed, err := sp.Sign(req.Data, key)
	if err != nil {
		return nil, err
	}

	return &SignResponse{
		KID:  key.ID().String(),
		Data: signed,
	}, nil
}

// Verify (RPC) ...
func (s *service) Verify(ctx context.Context, req *VerifyRequest) (*VerifyResponse, error) {
	sp := saltpack.NewSaltpack(s.ks)
	sp.SetArmored(req.Armored)
	verified, sender, err := sp.Verify(req.Data)
	if err != nil {
		return nil, err
	}
	return &VerifyResponse{Data: verified, KID: sender.String()}, nil
}

// SignStream (RPC) ...
func (s *service) SignStream(srv Keys_SignStreamServer) error {
	init := false

	var stream io.WriteCloser
	var buf bytes.Buffer

	ctx := srv.Context()
	var kid string
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
			key, err := s.parseKeyOrCurrent(req.KID)
			if err != nil {
				return err
			}
			kid = key.ID().String()

			sp := saltpack.NewSaltpack(s.ks)
			sp.SetArmored(req.Armored)
			s, streamErr := sp.NewSignStream(&buf, key, req.Detached)
			if streamErr != nil {
				return streamErr
			}
			stream = s
		} else {
			// Make sure request only sends data after init
			if req.KID != "" || req.Armored || req.Detached {
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
				if err := srv.Send(&SignStreamOutput{
					Data: out,
					KID:  kid,
				}); err != nil {
					return err
				}
				buf.Reset()
			}
		}
	}
	// Close stream and flush buffer
	stream.Close()
	if buf.Len() > 0 {
		out := buf.Bytes()
		if err := srv.Send(&SignStreamOutput{
			Data: out,
			KID:  kid,
		}); err != nil {
			return err
		}
		buf.Reset()
	}
	return nil
}

// VerifyStream (RPC) ...
func (s *service) VerifyStream(srv Keys_VerifyStreamServer) error {
	recvFn := func() ([]byte, error) {
		req, recvErr := srv.Recv()
		if recvErr != nil {
			return nil, recvErr
		}
		return req.Data, nil
	}

	reader := newStreamReader(srv.Context(), recvFn)
	sp := saltpack.NewSaltpack(s.ks)
	streamReader, sender, streamErr := sp.NewVerifyStream(reader)
	if streamErr != nil {
		return streamErr
	}
	sendFn := func(b []byte) error {
		resp := VerifyStreamOutput{
			KID:  sender.String(),
			Data: b,
		}
		return srv.Send(&resp)
	}
	return s.readFromStream(srv.Context(), streamReader, sendFn)
}

// VerifyArmoredStream (RPC) ...
func (s *service) VerifyArmoredStream(srv Keys_VerifyArmoredStreamServer) error {
	recvFn := func() ([]byte, error) {
		req, recvErr := srv.Recv()
		if recvErr != nil {
			return nil, recvErr
		}
		return req.Data, nil
	}

	reader := newStreamReader(srv.Context(), recvFn)
	sp := saltpack.NewSaltpack(s.ks)
	sp.SetArmored(true)
	streamReader, sender, streamErr := sp.NewVerifyStream(reader)
	if streamErr != nil {
		return streamErr
	}
	sendFn := func(b []byte) error {
		resp := VerifyStreamOutput{
			KID:  sender.String(),
			Data: b,
		}
		return srv.Send(&resp)
	}
	return s.readFromStream(srv.Context(), streamReader, sendFn)
}

// VerifyStreamClient ...
type VerifyStreamClient interface {
	Send(*VerifyStreamInput) error
	Recv() (*VerifyStreamOutput, error)
	grpc.ClientStream
}

// NewVerifyStreamClient ...
func NewVerifyStreamClient(ctx context.Context, cl KeysClient, armored bool) (VerifyStreamClient, error) {
	if armored {
		return cl.VerifyArmoredStream(ctx)
	}
	return cl.VerifyStream(ctx)
}

func (s *service) readFromStream(ctx context.Context, streamReader io.Reader, sendFn func(b []byte) error) error {
	buf := make([]byte, 1024*1024)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, err := streamReader.Read(buf)
		if n != 0 {
			if err := sendFn(buf[:n]); err != nil {
				return err
			}
		}
		if err != nil {
			if err == io.EOF {
				// Send empty bytes denotes EOF
				if err := sendFn([]byte{}); err != nil {
					return err
				}
				break
			}
			return err
		}

	}

	return nil
}
