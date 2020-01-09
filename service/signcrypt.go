package service

import (
	"bytes"
	"context"
	"io"
	strings "strings"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/saltpack"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Encrypt (RPC) data.
func (s *service) Encrypt(ctx context.Context, req *EncryptRequest) (*EncryptResponse, error) {
	if req.Recipients == "" {
		return nil, errors.Errorf("no recipients specified")
	}
	sender, senderErr := s.parseKey(req.Sender)
	if senderErr != nil {
		return nil, senderErr
	}

	recipients, err := keys.ParseIDs(strings.Split(req.Recipients, ","))
	if err != nil {
		return nil, err
	}

	sp := saltpack.NewSaltpack(s.ks)
	sp.SetArmored(req.Armored)
	data, err := sp.Signcrypt(req.Data, sender, recipients...)
	if err != nil {
		return nil, err
	}

	return &EncryptResponse{
		Data: data,
	}, nil
}

// Decrypt (RPC) data.
func (s *service) Decrypt(ctx context.Context, req *DecryptRequest) (*DecryptResponse, error) {
	sp := saltpack.NewSaltpack(s.ks)
	sp.SetArmored(req.Armored)
	logger.Debugf("Saltpack open (len=%d)", len(req.Data))
	decrypted, signer, err := sp.SigncryptOpen(req.Data)
	if err != nil {
		return nil, err
	}
	return &DecryptResponse{
		Data:   decrypted,
		Sender: signer.String(),
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
			sender, senderErr := s.parseKey(req.Sender)
			if senderErr != nil {
				return senderErr
			}
			if req.Recipients == "" {
				return errors.Errorf("no recipients specified")
			}
			recipients, err := keys.ParseIDs(strings.Split(req.Recipients, ","))
			if err != nil {
				return err
			}
			sp := saltpack.NewSaltpack(s.ks)
			sp.SetArmored(req.Armored)
			logger.Infof("Seal stream for %s from %s", req.Recipients, req.Sender)
			s, streamErr := sp.NewSigncryptStream(&buf, sender, recipients...)
			if streamErr != nil {
				return streamErr
			}
			stream = s
		} else {
			// Make sure request only sends data after init
			if req.Recipients != "" || req.Sender != "" || req.Armored {
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

// DecryptStreamClient ...
type DecryptStreamClient interface {
	Send(*DecryptStreamInput) error
	Recv() (*DecryptStreamOutput, error)
	grpc.ClientStream
}

// NewDecryptStreamClient ...
func NewDecryptStreamClient(ctx context.Context, cl KeysClient, armored bool) (DecryptStreamClient, error) {
	if armored {
		return cl.DecryptArmoredStream(ctx)
	}
	return cl.DecryptStream(ctx)
}

// DecryptStream (RPC) ...
func (s *service) DecryptStream(srv Keys_DecryptStreamServer) error {
	recvFn := func() ([]byte, error) {
		req, recvErr := srv.Recv()
		if recvErr != nil {
			return nil, recvErr
		}
		return req.Data, nil
	}

	reader := newStreamReader(srv.Context(), recvFn)
	sp := saltpack.NewSaltpack(s.ks)
	streamReader, signer, streamErr := sp.NewSigncryptOpenStream(reader)
	if streamErr != nil {
		return streamErr
	}
	sendFn := func(b []byte) error {
		resp := DecryptStreamOutput{
			Data:   b,
			Sender: signer.String(),
		}
		return srv.Send(&resp)
	}
	return s.readFromStream(srv.Context(), streamReader, sendFn)
}

// DecryptArmoredStream (RPC) ...
func (s *service) DecryptArmoredStream(srv Keys_DecryptArmoredStreamServer) error {
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
	streamReader, signer, streamErr := sp.NewSigncryptOpenStream(reader)
	if streamErr != nil {
		return streamErr
	}
	sendFn := func(b []byte) error {
		resp := DecryptStreamOutput{
			Data:   b,
			Sender: signer.String(),
		}
		return srv.Send(&resp)
	}
	return s.readFromStream(srv.Context(), streamReader, sendFn)
}
