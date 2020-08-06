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
	armored    bool
}

func (s *service) newEncrypt(ctx context.Context, recipients []string, sender string, mode EncryptMode, armored bool) (*encrypt, error) {
	if len(recipients) == 0 {
		return nil, errors.Errorf("no recipients specified")
	}

	recs, err := s.lookupAll(ctx, recipients, &LookupOpts{Verify: true})
	if err != nil {
		return nil, err
	}

	if mode == DefaultEncrypt {
		mode = SaltpackEncrypt
	}

	var kid keys.ID
	if sender != "" {
		s, err := s.lookup(ctx, sender, &LookupOpts{Verify: true})
		if err != nil {
			return nil, err
		}
		kid = s
	}

	return &encrypt{
		recipients: recs,
		sender:     kid,
		mode:       mode,
		armored:    armored,
	}, nil
}

// Encrypt (RPC) data.
func (s *service) Encrypt(ctx context.Context, req *EncryptRequest) (*EncryptResponse, error) {
	enc, err := s.newEncrypt(ctx, req.Recipients, req.Sender, req.Mode, req.Armored)
	if err != nil {
		return nil, err
	}

	var out []byte
	switch enc.mode {
	case SaltpackEncrypt:
		sbk, err := s.x25519Key(enc.sender)
		if err != nil {
			return nil, err
		}
		out, err = saltpack.Encrypt(req.Data, req.Armored, sbk, enc.recipients...)
		if err != nil {
			return nil, err
		}
	case SaltpackSigncrypt:
		if enc.sender == "" {
			return nil, errors.Errorf("no sender specified: sender is required for signcrypt mode")
		}
		sk, err := s.vault.EdX25519Key(enc.sender)
		if err != nil {
			return nil, err
		}
		if sk == nil {
			return nil, keys.NewErrNotFound(enc.sender.String())
		}
		out, err = saltpack.Signcrypt(req.Data, req.Armored, sk, enc.recipients...)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.Errorf("unsupported mode %s", enc.mode)
	}

	return &EncryptResponse{
		Data: out,
	}, nil
}

func (s *service) encryptWriteInOut(ctx context.Context, in string, out string, enc *encrypt) error {
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

	stream, err := s.encryptWriter(ctx, writer, enc)
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

func (s *service) encryptWriter(ctx context.Context, w io.Writer, enc *encrypt) (io.WriteCloser, error) {
	var stream io.WriteCloser
	switch enc.mode {
	case SaltpackEncrypt:
		sbk, err := s.x25519Key(enc.sender)
		if err != nil {
			return nil, err
		}
		logger.Infof("Encrypt stream for %s from %s", enc.recipients, enc.sender)
		stream, err = saltpack.NewEncryptStream(w, enc.armored, sbk, enc.recipients...)
		if err != nil {
			return nil, err
		}
	case SaltpackSigncrypt:
		if enc.sender == "" {
			return nil, errors.Errorf("no sender specified")
		}
		sk, err := s.vault.EdX25519Key(enc.sender)
		if err != nil {
			return nil, err
		}
		if sk == nil {
			return nil, keys.NewErrNotFound(enc.sender.String())
		}
		logger.Infof("Signcrypt stream for %s from %s", enc.recipients, enc.sender)
		stream, err = saltpack.NewSigncryptStream(w, enc.armored, sk, enc.recipients...)
		if err != nil {
			return nil, err
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
	in := req.In
	if in == "" {
		return errors.Errorf("in not specified")
	}
	out := req.Out
	if out == "" {
		out = in + ".enc"
	}

	enc, err := s.newEncrypt(srv.Context(), req.Recipients, req.Sender, req.Mode, req.Armored)
	if err != nil {
		return err
	}

	if err := s.encryptWriteInOut(srv.Context(), in, out, enc); err != nil {
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

			enc, err := s.newEncrypt(ctx, req.Recipients, req.Sender, req.Mode, req.Armored)
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
