package service

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/saltpack"
	"github.com/pkg/errors"
)

type sign struct {
	key      *keys.SignKey
	armored  bool
	detached bool
	sp       *saltpack.Saltpack
}

func (s *service) newSign(signer string, armored bool, detached bool) (*sign, error) {
	key, err := s.parseSigner(signer, true)
	if err != nil {
		return nil, err
	}
	sp := saltpack.NewSaltpack(s.ks)
	sp.SetArmored(armored)
	return &sign{
		key:      key,
		armored:  armored,
		detached: detached,
		sp:       sp,
	}, nil
}

// Sign (RPC) ...
func (s *service) Sign(ctx context.Context, req *SignRequest) (*SignResponse, error) {
	sign, err := s.newSign(req.Signer, req.Armored, req.Detached)
	if err != nil {
		return nil, err
	}

	signed, err := sign.sp.Sign(req.Data, sign.key)
	if err != nil {
		return nil, err
	}

	return &SignResponse{
		Data: signed,
		KID:  sign.key.ID().String(),
	}, nil
}

// SignFile (RPC) ...
func (s *service) SignFile(srv Keys_SignFileServer) error {
	req, err := srv.Recv()
	if err != nil {
		return err
	}
	if req.In == "" {
		return errors.Errorf("in not specified")
	}
	if req.Out == "" {
		return errors.Errorf("out not specified")
	}

	sign, err := s.newSign(req.Signer, req.Armored, req.Detached)
	if err != nil {
		return err
	}

	if err := s.signWriteInOut(srv.Context(), req.In, req.Out, sign); err != nil {
		return err
	}

	if err := srv.Send(&SignFileOutput{
		KID: sign.key.ID().String(),
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
			sign, err := s.newSign(req.Signer, req.Armored, req.Detached)
			if err != nil {
				return err
			}
			s, err := s.signWriter(ctx, &buf, sign)
			if err != nil {
				return err
			}
			stream = s
			kid = sign.key.ID()

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
	stream.Close()
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

func (s *service) signWriter(ctx context.Context, w io.Writer, sign *sign) (io.WriteCloser, error) {
	return sign.sp.NewSignStream(w, sign.key, sign.detached)
}

func (s *service) signWriteInOut(ctx context.Context, in string, out string, sign *sign) error {
	outTmp := out + ".tmp"
	defer os.Remove(outTmp)
	outFile, err := os.Create(outTmp)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(outFile)

	stream, err := s.signWriter(ctx, writer, sign)
	if err != nil {
		return err
	}

	inFile, err := os.Open(in)
	if err != nil {
		return err
	}
	reader := bufio.NewReader(inFile)
	if _, err := reader.WriteTo(stream); err != nil {
		return err
	}

	stream.Close()
	writer.Flush()

	if err := os.Rename(outTmp, out); err != nil {
		return err
	}

	return nil
}
