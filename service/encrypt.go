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

type encrypt struct {
	recipients []keys.ID
	sender     keys.ID
	mode       EncryptMode
}

func (s *service) newEncrypt(ctx context.Context, recipients []string, sender string, mode EncryptMode) (*encrypt, error) {
	if len(recipients) == 0 {
		return nil, errors.Errorf("no recipients specified")
	}

	identities, err := s.parseIdentities(ctx, recipients)
	if err != nil {
		return nil, err
	}

	if mode == DefaultEncryptMode {
		mode = EncryptV2
	}

	var kid keys.ID
	if sender != "" {
		s, err := s.parseIdentity(ctx, sender)
		if err != nil {
			return nil, err
		}
		kid = s
	}

	return &encrypt{
		recipients: identities,
		sender:     kid,
		mode:       mode,
	}, nil
}

// Encrypt (RPC) data.
func (s *service) Encrypt(ctx context.Context, req *EncryptRequest) (*EncryptResponse, error) {
	enc, err := s.newEncrypt(ctx, req.Recipients, req.Sender, req.Mode)
	if err != nil {
		return nil, err
	}

	sp := saltpack.NewSaltpack(s.ks)
	var out []byte
	switch enc.mode {
	case EncryptV2:
		sbk, err := s.parseBoxKey(enc.sender)
		if err != nil {
			return nil, err
		}
		if req.Armored {
			data, err := sp.EncryptArmored(req.Data, sbk, enc.recipients...)
			if err != nil {
				return nil, err
			}
			out = []byte(data)
		} else {
			data, err := sp.Encrypt(req.Data, sbk, enc.recipients...)
			if err != nil {
				return nil, err
			}
			out = data
		}
	case SigncryptV1:
		if enc.sender == "" {
			return nil, errors.Errorf("no sender specified: sender is required for signcrypt mode")
		}
		sk, err := s.ks.EdX25519Key(enc.sender)
		if err != nil {
			return nil, err
		}
		if sk == nil {
			return nil, keys.NewErrNotFound(enc.sender.String())
		}
		if req.Armored {
			data, err := sp.SigncryptArmored(req.Data, sk, enc.recipients...)
			if err != nil {
				return nil, err
			}
			out = []byte(data)
		} else {
			data, err := sp.Signcrypt(req.Data, sk, enc.recipients...)
			if err != nil {
				return nil, err
			}
			out = data
		}
	default:
		return nil, errors.Errorf("unsupported mode %s", enc.mode)
	}

	return &EncryptResponse{
		Data: out,
	}, nil
}

func (s *service) encryptWriteInOut(ctx context.Context, in string, out string, enc *encrypt, armored bool) error {
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

	stream, err := s.encryptWriter(ctx, writer, enc, armored)
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

func (s *service) encryptWriter(ctx context.Context, w io.Writer, enc *encrypt, armored bool) (io.WriteCloser, error) {
	var stream io.WriteCloser

	sp := saltpack.NewSaltpack(s.ks)
	switch enc.mode {
	case EncryptV2:
		sbk, err := s.parseBoxKey(enc.sender)
		if err != nil {
			return nil, err
		}
		logger.Infof("Encrypt stream for %s from %s", enc.recipients, enc.sender)
		if armored {
			s, err := sp.NewEncryptArmoredStream(w, sbk, enc.recipients...)
			if err != nil {
				return nil, err
			}
			stream = s
		} else {
			s, err := sp.NewEncryptStream(w, sbk, enc.recipients...)
			if err != nil {
				return nil, err
			}
			stream = s
		}

	case SigncryptV1:
		if enc.sender == "" {
			return nil, errors.Errorf("no sender specified")
		}
		sk, err := s.ks.EdX25519Key(enc.sender)
		if err != nil {
			return nil, err
		}
		if sk == nil {
			return nil, keys.NewErrNotFound(enc.sender.String())
		}
		logger.Infof("Signcrypt stream for %s from %s", enc.recipients, enc.sender)
		if armored {
			s, err := sp.NewSigncryptArmoredStream(w, sk, enc.recipients...)
			if err != nil {
				return nil, err
			}
			stream = s
		} else {
			s, err := sp.NewSigncryptStream(w, sk, enc.recipients...)
			if err != nil {
				return nil, err
			}
			stream = s
		}
	default:
		return nil, errors.Errorf("unsupported mode %s", enc.mode)
	}
	return stream, nil
}

// EncryptFile (RPC) ...
func (s *service) EncryptFile(srv Keys_EncryptFileServer) error {
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

	enc, err := s.newEncrypt(srv.Context(), req.Recipients, req.Sender, req.Mode)
	if err != nil {
		return err
	}

	if err := s.encryptWriteInOut(srv.Context(), req.In, req.Out, enc, req.Armored); err != nil {
		return err
	}

	if err := srv.Send(&EncryptFileOutput{}); err != nil {
		return err
	}

	return nil
}

// EncryptStream (RPC) ...
func (s *service) EncryptStream(srv Keys_EncryptStreamServer) error {
	init := false

	var stream io.WriteCloser
	var buf bytes.Buffer

	ctx := srv.Context()
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

			enc, err := s.newEncrypt(ctx, req.Recipients, req.Sender, req.Mode)
			if err != nil {
				return err
			}

			s, err := s.encryptWriter(ctx, &buf, enc, req.Armored)
			if err != nil {
				return err
			}
			stream = s

		} else {
			// Make sure request only sends data after init
			if len(req.Recipients) != 0 || req.Sender != "" {
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
				if err := srv.Send(&EncryptOutput{Data: out}); err != nil {
					return err
				}
				buf.Reset()
			}
		}
	}
	logger.Debugf("Stream close")
	// Close stream and flush buffer
	if err := stream.Close(); err != nil {
		return err
	}
	if buf.Len() > 0 {
		out := buf.Bytes()
		if err := srv.Send(&EncryptOutput{Data: out}); err != nil {
			return err
		}
		buf.Reset()
	}
	return nil
}
