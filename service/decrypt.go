package service

import (
	"bufio"
	"context"
	"io"
	"os"
	strings "strings"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/saltpack"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func (s *service) decryptSender(ctx context.Context, kid keys.ID, mode EncryptMode) (*Key, error) {
	if kid == "" {
		logger.Infof("No decrypt sender")
		return nil, nil
	}
	// If EncryptV2 check the kid.
	if mode == EncryptV2 {
		k, err := s.checkSenderID(kid)
		if err != nil {
			return nil, err
		}
		kid = k
	}
	return s.loadKey(ctx, kid)
}

// Decrypt (RPC) data.
func (s *service) Decrypt(ctx context.Context, req *DecryptRequest) (*DecryptResponse, error) {
	var decrypted []byte
	var kid keys.ID
	var err error
	sp := saltpack.NewSaltpack(s.ks)

	mode := req.Mode
	if mode == DefaultEncryptMode {
		mode = EncryptV2
	}

	switch mode {
	case EncryptV2:
		if req.Armored {
			decrypted, kid, err = sp.DecryptArmored(string(req.Data))
		} else {
			decrypted, kid, err = sp.Decrypt(req.Data)
		}
	case SigncryptV1:
		if req.Armored {
			decrypted, kid, err = sp.SigncryptArmoredOpen(string(req.Data))
		} else {
			decrypted, kid, err = sp.SigncryptOpen(req.Data)
		}
	default:
		return nil, errors.Errorf("unsupported mode %s", mode)
	}

	if err != nil {
		if err.Error() == "failed to read header bytes" {
			return nil, errors.Errorf("invalid data")
		}
		return nil, err
	}

	sender, err := s.decryptSender(ctx, kid, mode)
	if err != nil {
		return nil, err
	}

	return &DecryptResponse{
		Data:   decrypted,
		Sender: sender,
	}, nil
}

// DecryptFile (RPC) ...
func (s *service) DecryptFile(srv Keys_DecryptFileServer) error {
	req, err := srv.Recv()
	if err != nil {
		return err
	}
	if req.In == "" {
		return errors.Errorf("in not specified")
	}
	out := req.Out
	if out == "" {
		if strings.HasSuffix(req.In, ".enc") {
			out = strings.TrimSuffix(req.In, ".enc")
		}
		out, err = nextPath(out)
		if err != nil {
			return err
		}
	}

	sender, err := s.decryptWriteInOut(srv.Context(), req.In, out, req.Mode, req.Armored)
	if err != nil {
		return err
	}

	if err := srv.Send(&DecryptFileOutput{
		Sender: sender,
		Out:    out,
	}); err != nil {
		return err
	}

	return nil
}

// DecryptStreamClient can send and recieve input and output.
type DecryptStreamClient interface {
	Send(*DecryptInput) error
	Recv() (*DecryptOutput, error)
	grpc.ClientStream
}

// NewDecryptStreamClient returns DecryptStreamClient based on options.
func NewDecryptStreamClient(ctx context.Context, cl KeysClient, armored bool, mode EncryptMode) (DecryptStreamClient, error) {
	switch mode {
	case DefaultEncryptMode, EncryptV2:
		if armored {
			return cl.DecryptArmoredStream(ctx)
		}
		return cl.DecryptStream(ctx)
	case SigncryptV1:
		if armored {
			return cl.SigncryptOpenArmoredStream(ctx)
		}
		return cl.SigncryptOpenStream(ctx)
	default:
		return nil, errors.Errorf("unsupported mode %s", mode)
	}
}

// DecryptStream (RPC) ...
func (s *service) DecryptStream(srv Keys_DecryptStreamServer) error {
	return s.decryptStream(srv, EncryptV2, false)
}

// DecryptArmoredStream (RPC) ...
func (s *service) DecryptArmoredStream(srv Keys_DecryptArmoredStreamServer) error {
	return s.decryptStream(srv, EncryptV2, true)
}

// SigncryptOpenStream (RPC) ...
func (s *service) SigncryptOpenStream(srv Keys_SigncryptOpenStreamServer) error {
	return s.decryptStream(srv, SigncryptV1, false)
}

// SigncryptOpenArmoredStream (RPC) ...
func (s *service) SigncryptOpenArmoredStream(srv Keys_SigncryptOpenArmoredStreamServer) error {
	return s.decryptStream(srv, SigncryptV1, true)
}

type decryptStreamServer interface {
	Send(*DecryptOutput) error
	Recv() (*DecryptInput, error)
	grpc.ServerStream
}

func (s *service) decryptStream(srv decryptStreamServer, mode EncryptMode, armored bool) error {
	recvFn := func() ([]byte, error) {
		req, recvErr := srv.Recv()
		if recvErr != nil {
			return nil, recvErr
		}
		return req.Data, nil
	}

	reader := newStreamReader(srv.Context(), recvFn)

	streamReader, kid, err := s.decryptReader(srv.Context(), reader, mode, armored)
	if err != nil {
		return err
	}

	sender, err := s.decryptSender(srv.Context(), kid, mode)
	if err != nil {
		return err
	}

	sendFn := func(b []byte, sender *Key) error {
		resp := DecryptOutput{
			Data:   b,
			Sender: sender,
		}
		return srv.Send(&resp)
	}
	return s.readFromStream(srv.Context(), streamReader, sender, sendFn)
}

func (s *service) decryptReader(ctx context.Context, reader io.Reader, mode EncryptMode, armored bool) (io.Reader, keys.ID, error) {
	sp := saltpack.NewSaltpack(s.ks)
	switch mode {
	case DefaultEncryptMode, EncryptV2:
		if armored {
			return sp.NewDecryptArmoredStream(reader)
		}
		return sp.NewDecryptStream(reader)
	case SigncryptV1:
		if armored {
			return sp.NewSigncryptArmoredOpenStream(reader)
		}
		return sp.NewSigncryptOpenStream(reader)
	default:
		return nil, "", errors.Errorf("unsupported mode %s", mode)
	}
}

func (s *service) decryptWriteInOut(ctx context.Context, in string, out string, mode EncryptMode, armored bool) (*Key, error) {
	inFile, err := os.Open(in) // #nosec
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(inFile)

	decReader, kid, err := s.decryptReader(ctx, reader, mode, armored)
	if err != nil {
		return nil, err
	}
	outTmp := out + ".tmp"
	defer os.Remove(outTmp)
	outFile, err := os.Create(outTmp)
	if err != nil {
		return nil, err
	}
	writer := bufio.NewWriter(outFile)

	if _, err := writer.ReadFrom(decReader); err != nil {
		return nil, err
	}
	if err := writer.Flush(); err != nil {
		return nil, err
	}

	if err := os.Rename(outTmp, out); err != nil {
		return nil, err
	}

	sender, err := s.decryptSender(ctx, kid, mode)
	if err != nil {
		return nil, err
	}

	return sender, nil
}
