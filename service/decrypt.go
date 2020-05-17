package service

import (
	"bufio"
	"context"
	"io"
	"os"
	"strings"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/saltpack"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func (s *service) findSender(ctx context.Context, kid keys.ID) (*Key, error) {
	if kid == "" {
		logger.Infof("No decrypt sender")
		return nil, nil
	}
	k, err := s.convertX25519ID(kid)
	if err != nil {
		return nil, err
	}
	kid = k
	return s.loadKey(ctx, kid)
}

// Decrypt (RPC) data.
func (s *service) Decrypt(ctx context.Context, req *DecryptRequest) (*DecryptResponse, error) {
	mode := req.Mode
	if mode == DefaultEncryptMode {
		mode = EncryptV2
	}

	// TODO: Autodetect if input data is armored or not

	var decrypted []byte
	var kid keys.ID
	var decryptErr error
	switch mode {
	case EncryptV2:
		var sender *keys.X25519PublicKey
		if req.Armored {
			decrypted, sender, decryptErr = saltpack.DecryptArmored(string(req.Data), s.ks)
			if sender != nil {
				kid = sender.ID()
			}
		} else {
			decrypted, sender, decryptErr = saltpack.Decrypt(req.Data, s.ks)
			if sender != nil {
				kid = sender.ID()
			}
		}
	case SigncryptV1:
		var sender *keys.EdX25519PublicKey
		if req.Armored {
			decrypted, sender, decryptErr = saltpack.SigncryptArmoredOpen(string(req.Data), s.ks)
			if sender != nil {
				kid = sender.ID()
			}
		} else {
			decrypted, sender, decryptErr = saltpack.SigncryptOpen(req.Data, s.ks)
			if sender != nil {
				kid = sender.ID()
			}
		}
	default:
		return nil, errors.Errorf("unsupported mode %s", mode)
	}

	if decryptErr != nil {
		if decryptErr.Error() == "failed to read header bytes" {
			return nil, errors.Errorf("invalid data")
		}
		return nil, decryptErr
	}

	senderKey, err := s.findSender(ctx, kid)
	if err != nil {
		return nil, err
	}

	return &DecryptResponse{
		Data:   decrypted,
		Sender: senderKey,
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
	}
	exists, err := fileExists(out)
	if err != nil {
		return err
	}
	if exists {
		return errors.Errorf("file already exists %s", out)
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

	sender, err := s.findSender(srv.Context(), kid)
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
	var out io.Reader
	var kid keys.ID
	var decryptErr error
	switch mode {
	case DefaultEncryptMode, EncryptV2:
		var sender *keys.X25519PublicKey
		if armored {
			out, sender, decryptErr = saltpack.NewDecryptArmoredStream(reader, s.ks)
			if sender != nil {
				kid = sender.ID()
			}
		} else {
			out, sender, decryptErr = saltpack.NewDecryptStream(reader, s.ks)
			if sender != nil {
				kid = sender.ID()
			}
		}
	case SigncryptV1:
		var sender *keys.EdX25519PublicKey
		if armored {
			out, sender, decryptErr = saltpack.NewSigncryptArmoredOpenStream(reader, s.ks)
			if sender != nil {
				kid = sender.ID()
			}
		} else {
			out, sender, decryptErr = saltpack.NewSigncryptOpenStream(reader, s.ks)
			if sender != nil {
				kid = sender.ID()
			}
		}
	default:
		return nil, "", errors.Errorf("unsupported mode %s", mode)
	}

	return out, kid, decryptErr
}

func (s *service) decryptWriteInOut(ctx context.Context, in string, out string, mode EncryptMode, armored bool) (*Key, error) {
	inFile, err := os.Open(in) // #nosec
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = inFile.Close()
	}()
	reader := bufio.NewReader(inFile)

	decReader, kid, err := s.decryptReader(ctx, reader, mode, armored)
	if err != nil {
		return nil, err
	}
	outTmp := out + ".tmp"
	outFile, err := os.Create(outTmp)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = outFile.Close()
		_ = os.Remove(outTmp)
	}()

	writer := bufio.NewWriter(outFile)

	if _, err := writer.ReadFrom(decReader); err != nil {
		return nil, err
	}
	if err := writer.Flush(); err != nil {
		return nil, err
	}
	if err := inFile.Close(); err != nil {
		return nil, err
	}
	if err := outFile.Close(); err != nil {
		return nil, err
	}

	if err := os.Rename(outTmp, out); err != nil {
		return nil, err
	}

	sender, err := s.findSender(ctx, kid)
	if err != nil {
		return nil, err
	}

	return sender, nil
}
