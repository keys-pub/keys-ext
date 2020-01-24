package service

import (
	"bytes"
	"context"
	"io"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/saltpack"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Encrypt (RPC) data.
func (s *service) Encrypt(ctx context.Context, req *EncryptRequest) (*EncryptResponse, error) {
	if len(req.Recipients) == 0 {
		return nil, errors.Errorf("no recipients specified")
	}

	recipients, err := keys.ParseIDs(req.Recipients)
	if err != nil {
		return nil, err
	}

	sp := saltpack.NewSaltpack(s.ks)
	sp.SetArmored(req.Armored)

	var out []byte
	switch req.Mode {
	case EncryptV2, DefaultEncryptMode:
		sender, err := s.parseBoxKey(req.Sender, false)
		if err != nil {
			return nil, err
		}
		data, err := sp.Encrypt(req.Data, sender, recipients...)
		if err != nil {
			return nil, err
		}
		out = data
	case SigncryptV1:
		if req.Sender == "" {
			return nil, errors.Errorf("no sender specified: sender is required for signcrypt mode")
		}
		sender, err := s.parseSignKey(req.Sender, true)
		if err != nil {
			return nil, err
		}
		data, err := sp.Signcrypt(req.Data, sender, recipients...)
		if err != nil {
			return nil, err
		}
		out = data
	default:
		return nil, errors.Errorf("unsupported mode %s", req.Mode)
	}

	return &EncryptResponse{
		Data: out,
	}, nil
}

// Decrypt (RPC) data.
func (s *service) Decrypt(ctx context.Context, req *DecryptRequest) (*DecryptResponse, error) {
	sp := saltpack.NewSaltpack(s.ks)
	sp.SetArmored(req.Armored)
	logger.Debugf("Saltpack open (len=%d)", len(req.Data))

	var decrypted []byte
	var sender keys.ID
	switch req.Mode {
	case EncryptV2, DefaultEncryptMode:
		d, s, err := sp.Decrypt(req.Data)
		if err != nil {
			return nil, err
		}
		decrypted, sender = d, s
	case SigncryptV1:
		d, s, err := sp.SigncryptOpen(req.Data)
		if err != nil {
			return nil, err
		}
		decrypted, sender = d, s
	default:
		return nil, errors.Errorf("unsupported mode %s", req.Mode)
	}

	return &DecryptResponse{
		Data:   decrypted,
		Sender: sender.String(),
	}, nil
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
			recipients, err := keys.ParseIDs(req.Recipients)
			if err != nil {
				return err
			}
			if len(req.Recipients) == 0 {
				return errors.Errorf("no recipients specified")
			}
			sp := saltpack.NewSaltpack(s.ks)
			sp.SetArmored(req.Armored)

			switch req.Mode {
			case EncryptV2, DefaultEncryptMode:
				sender, err := s.parseBoxKey(req.Sender, false)
				if err != nil {
					return err
				}
				logger.Infof("Encrypt stream for %s from %s", req.Recipients, req.Sender)
				s, err := sp.NewEncryptStream(&buf, sender, recipients...)
				if err != nil {
					return err
				}
				stream = s
			case SigncryptV1:
				if req.Sender == "" {
					return errors.Errorf("no sender specified")
				}
				sender, err := s.parseSignKey(req.Sender, true)
				if err != nil {
					return err
				}
				logger.Infof("Signcrypt stream for %s from %s", req.Recipients, req.Sender)
				s, err := sp.NewSigncryptStream(&buf, sender, recipients...)
				if err != nil {
					return err
				}
				stream = s
			}
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
				if err := srv.Send(&EncryptStreamOutput{Data: out}); err != nil {
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
		if err := srv.Send(&EncryptStreamOutput{Data: out}); err != nil {
			return err
		}
		buf.Reset()
	}
	return nil
}

// DecryptStreamClient can send and recieve input and output.
type DecryptStreamClient interface {
	Send(*DecryptStreamInput) error
	Recv() (*DecryptStreamOutput, error)
	grpc.ClientStream
}

// NewDecryptStreamClient returns DecryptStreamClient based on options.
func NewDecryptStreamClient(ctx context.Context, cl KeysClient, armored bool, mode EncryptMode) (DecryptStreamClient, error) {
	switch mode {
	case DefaultEncryptMode, EncryptV2:
		if armored {
			return cl.SigncryptOpenArmoredStream(ctx)
		}
		return cl.SigncryptOpenStream(ctx)
	case SigncryptV1:
		if armored {
			return cl.DecryptArmoredStream(ctx)
		}
		return cl.DecryptStream(ctx)
	default:
		return nil, errors.Errorf("unsupported mode %s", mode)
	}
}

// DecryptStream (RPC) ...
func (s *service) DecryptStream(srv Keys_DecryptStreamServer) error {
	return s.decryptStream(srv, false, EncryptV2)
}

// DecryptArmoredStream (RPC) ...
func (s *service) DecryptArmoredStream(srv Keys_DecryptArmoredStreamServer) error {
	return s.decryptStream(srv, true, EncryptV2)
}

// SigncryptOpenStream (RPC) ...
func (s *service) SigncryptOpenStream(srv Keys_SigncryptOpenStreamServer) error {
	return s.decryptStream(srv, false, SigncryptV1)
}

// SigncryptOpenArmoredStream (RPC) ...
func (s *service) SigncryptOpenArmoredStream(srv Keys_SigncryptOpenArmoredStreamServer) error {
	return s.decryptStream(srv, true, SigncryptV1)
}

type decryptStreamServer interface {
	Send(*DecryptStreamOutput) error
	Recv() (*DecryptStreamInput, error)
	grpc.ServerStream
}

func (s *service) decryptStream(srv decryptStreamServer, armored bool, mode EncryptMode) error {
	recvFn := func() ([]byte, error) {
		req, recvErr := srv.Recv()
		if recvErr != nil {
			return nil, recvErr
		}
		return req.Data, nil
	}

	reader := newStreamReader(srv.Context(), recvFn)
	sp := saltpack.NewSaltpack(s.ks)
	sp.SetArmored(armored)
	var streamReader io.Reader
	var sender keys.ID
	switch mode {
	case EncryptV2, DefaultEncryptMode:
		r, s, err := sp.NewDecryptStream(reader)
		if err != nil {
			return err
		}
		streamReader, sender = r, s
	case SigncryptV1:
		r, s, err := sp.NewSigncryptOpenStream(reader)
		if err != nil {
			return err
		}
		streamReader, sender = r, s
	}
	sendFn := func(b []byte) error {
		resp := DecryptStreamOutput{
			Data:   b,
			Sender: sender.String(),
		}
		return srv.Send(&resp)
	}
	return s.readFromStream(srv.Context(), streamReader, sendFn)
}
