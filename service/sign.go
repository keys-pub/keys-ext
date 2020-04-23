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

// Sign (RPC) ...
func (s *service) Sign(ctx context.Context, req *SignRequest) (*SignResponse, error) {
	key, err := s.parseSigner(req.Signer, true)
	if err != nil {
		return nil, err
	}

	sp := saltpack.NewSaltpack(s.ks)
	var signed []byte
	if req.Armored {
		s, err := sp.SignArmored(req.Data, key)
		if err != nil {
			return nil, err
		}
		signed = []byte(s)
	} else {
		s, err := sp.Sign(req.Data, key)
		if err != nil {
			return nil, err
		}
		signed = s
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
		out = in + ".sig"
	}

	key, err := s.parseSigner(req.Signer, true)
	if err != nil {
		return err
	}

	if err := s.signWriteInOut(srv.Context(), in, out, key, req.Armored, req.Detached); err != nil {
		return err
	}

	if err := srv.Send(&SignFileOutput{
		KID: key.ID().String(),
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
			s, err := s.signWriter(ctx, &buf, key, req.Armored, req.Detached)
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

func (s *service) signWriter(ctx context.Context, w io.Writer, key *keys.EdX25519Key, armored bool, detached bool) (io.WriteCloser, error) {
	sp := saltpack.NewSaltpack(s.ks)
	if armored {
		if detached {
			logger.Debugf("Signing mode: armored/detached")
			return sp.NewSignArmoredDetachedStream(w, key)
		}
		logger.Debugf("Signing mode: armored")
		return sp.NewSignArmoredStream(w, key)
	}
	if detached {
		logger.Debugf("Signing mode: detached")
		return sp.NewSignDetachedStream(w, key)
	}
	logger.Debugf("Signing mode: default")
	return sp.NewSignStream(w, key)
}

func (s *service) signWriteInOut(ctx context.Context, in string, out string, key *keys.EdX25519Key, armored bool, detached bool) error {
	logger.Infof("Signing %s to %s", in, out)

	outTmp := out + ".tmp"
	outFile, err := os.Create(outTmp)
	if err != nil {
		return err
	}
	defer func() {
		_ = outFile.Close()
		_ = os.Remove(outTmp)
	}()
	writer := bufio.NewWriter(outFile)

	stream, err := s.signWriter(ctx, writer, key, armored, detached)
	if err != nil {
		return err
	}

	inFile, err := os.Open(in) // #nosec
	if err != nil {
		return err
	}
	defer func() {
		_ = inFile.Close()
	}()
	reader := bufio.NewReader(inFile)
	if _, err := reader.WriteTo(stream); err != nil {
		return err
	}

	if err := stream.Close(); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	if err := inFile.Close(); err != nil {
		return err
	}
	if err := outFile.Close(); err != nil {
		return err
	}

	if err := os.Rename(outTmp, out); err != nil {
		return err
	}

	return nil
}
