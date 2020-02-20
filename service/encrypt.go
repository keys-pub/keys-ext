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
	signer     keys.ID
	mode       EncryptMode
	sp         *saltpack.Saltpack
}

func (s *service) newEncrypt(ctx context.Context, recipients []string, signer string, armored bool, mode EncryptMode) (*encrypt, error) {
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
	if signer != "" {
		s, err := s.parseIdentity(ctx, signer)
		if err != nil {
			return nil, err
		}
		kid = s
	}

	sp := saltpack.NewSaltpack(s.ks)
	sp.SetArmored(armored)

	return &encrypt{
		recipients: identities,
		signer:     kid,
		mode:       mode,
		sp:         sp,
	}, nil
}

// Encrypt (RPC) data.
func (s *service) Encrypt(ctx context.Context, req *EncryptRequest) (*EncryptResponse, error) {
	enc, err := s.newEncrypt(ctx, req.Recipients, req.Signer, req.Armored, req.Mode)
	if err != nil {
		return nil, err
	}

	var out []byte
	switch enc.mode {
	case EncryptV2:
		sbk, err := s.parseBoxKey(enc.signer)
		if err != nil {
			return nil, err
		}
		data, err := enc.sp.Encrypt(req.Data, sbk, enc.recipients...)
		if err != nil {
			return nil, err
		}
		out = data
	case SigncryptV1:
		if enc.signer == "" {
			return nil, errors.Errorf("no signer specified: signer is required for signcrypt mode")
		}
		sk, err := s.ks.EdX25519Key(enc.signer)
		if err != nil {
			return nil, err
		}
		if sk == nil {
			return nil, keys.NewErrNotFound(enc.signer.String())
		}
		data, err := enc.sp.Signcrypt(req.Data, sk, enc.recipients...)
		if err != nil {
			return nil, err
		}
		out = data
	default:
		return nil, errors.Errorf("unsupported mode %s", enc.mode)
	}

	return &EncryptResponse{
		Data: out,
	}, nil
}

func (s *service) encryptWriteInOut(ctx context.Context, in string, out string, enc *encrypt) error {
	outTmp := out + ".tmp"
	defer os.Remove(outTmp)
	outFile, err := os.Create(outTmp)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(outFile)

	stream, err := s.encryptWriter(ctx, writer, enc)
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

func (s *service) encryptWriter(ctx context.Context, w io.Writer, enc *encrypt) (io.WriteCloser, error) {
	var stream io.WriteCloser

	switch enc.mode {
	case EncryptV2:
		sbk, err := s.parseBoxKey(enc.signer)
		if err != nil {
			return nil, err
		}
		logger.Infof("Encrypt stream for %s from %s", enc.recipients, enc.signer)
		s, err := enc.sp.NewEncryptStream(w, sbk, enc.recipients...)
		if err != nil {
			return nil, err
		}
		stream = s
	case SigncryptV1:
		if enc.signer == "" {
			return nil, errors.Errorf("no signer specified")
		}
		sk, err := s.ks.EdX25519Key(enc.signer)
		if err != nil {
			return nil, err
		}
		if sk == nil {
			return nil, keys.NewErrNotFound(enc.signer.String())
		}
		logger.Infof("Signcrypt stream for %s from %s", enc.recipients, enc.signer)
		s, err := enc.sp.NewSigncryptStream(w, sk, enc.recipients...)
		if err != nil {
			return nil, err
		}
		stream = s
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

	enc, err := s.newEncrypt(srv.Context(), req.Recipients, req.Signer, req.Armored, req.Mode)
	if err != nil {
		return err
	}

	if err := s.encryptWriteInOut(srv.Context(), req.In, req.Out, enc); err != nil {
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

			enc, err := s.newEncrypt(ctx, req.Recipients, req.Signer, req.Armored, req.Mode)
			if err != nil {
				return err
			}

			s, err := s.encryptWriter(ctx, &buf, enc)
			if err != nil {
				return err
			}
			stream = s

		} else {
			// Make sure request only sends data after init
			if len(req.Recipients) != 0 || req.Signer != "" {
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
	stream.Close()
	if buf.Len() > 0 {
		out := buf.Bytes()
		if err := srv.Send(&EncryptOutput{Data: out}); err != nil {
			return err
		}
		buf.Reset()
	}
	return nil
}
