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
	k, err := s.convertIfX25519ID(kid)
	if err != nil {
		return nil, err
	}
	kid = k
	return s.loadKey(ctx, kid)
}

// Decrypt (RPC) data.
func (s *service) Decrypt(ctx context.Context, req *DecryptRequest) (*DecryptResponse, error) {
	out, key, enc, err := saltpack.Open(req.Data, s.vault)
	if err != nil {
		if err.Error() == "failed to read header bytes" {
			return nil, errors.Errorf("invalid data")
		}
		return nil, err
	}
	mode, err := modeFromEncoding(enc)
	if err != nil {
		return nil, err
	}

	var sender *Key
	if key != nil {
		s, err := s.findSender(ctx, key.ID())
		if err != nil {
			return nil, err
		}
		sender = s
	}

	return &DecryptResponse{
		Data:   out,
		Sender: sender,
		Mode:   mode,
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
		} else {
			out = req.In + ".dec"
		}
	}
	exists, err := pathExists(out)
	if err != nil {
		return err
	}
	if exists {
		return errors.Errorf("file already exists %s", out)
	}

	sender, mode, err := s.decryptWriteInOut(srv.Context(), req.In, out)
	if err != nil {
		return errors.Wrapf(err, "failed to decrypt")
	}

	if err := srv.Send(&DecryptFileOutput{
		Sender: sender,
		Out:    out,
		Mode:   mode,
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
func NewDecryptStreamClient(ctx context.Context, cl KeysClient) (DecryptStreamClient, error) {
	return cl.DecryptStream(ctx)
	// return cl.DecryptArmoredStream(ctx)
	// return cl.SigncryptOpenStream(ctx)
	// return cl.SigncryptOpenArmoredStream(ctx)
}

// DecryptStream (RPC) ...
func (s *service) DecryptStream(srv Keys_DecryptStreamServer) error {
	return s.decryptStream(srv)
}

type decryptStreamServer interface {
	Send(*DecryptOutput) error
	Recv() (*DecryptInput, error)
	grpc.ServerStream
}

func (s *service) decryptStream(srv decryptStreamServer) error {
	recvFn := func() ([]byte, error) {
		req, recvErr := srv.Recv()
		if recvErr != nil {
			return nil, recvErr
		}
		return req.Data, nil
	}

	reader := newStreamReader(srv.Context(), recvFn)

	out, key, enc, err := saltpack.NewReader(reader, s.vault)
	if err != nil {
		return err
	}
	mode, err := modeFromEncoding(enc)
	if err != nil {
		return err
	}

	var sender *Key
	if key != nil {
		s, err := s.findSender(srv.Context(), key.ID())
		if err != nil {
			return err
		}
		sender = s
	}

	sendFn := func(b []byte, sender *Key) error {
		resp := DecryptOutput{
			Data:   b,
			Sender: sender,
			Mode:   mode,
		}
		return srv.Send(&resp)
	}
	return s.readFromStream(srv.Context(), out, sender, sendFn)
}

func modeFromEncoding(enc saltpack.Encoding) (EncryptMode, error) {
	switch enc {
	case saltpack.EncryptEncoding:
		return SaltpackEncrypt, nil
	case saltpack.SigncryptEncoding:
		return SaltpackSigncrypt, nil
	default:
		return DefaultEncrypt, errors.Errorf("invalid encoding %q", enc)
	}
}

func (s *service) decryptWriteInOut(ctx context.Context, in string, out string) (*Key, EncryptMode, error) {
	inFile, err := os.Open(in) // #nosec
	if err != nil {
		return nil, DefaultEncrypt, errors.Wrapf(err, "failed to open %s", in)
	}
	defer func() {
		_ = inFile.Close()
	}()
	reader := bufio.NewReader(inFile)

	decReader, key, enc, err := saltpack.NewReader(reader, s.vault)
	if err != nil {
		return nil, DefaultEncrypt, err
	}
	mode, err := modeFromEncoding(enc)
	if err != nil {
		return nil, DefaultEncrypt, err
	}

	if err := writeFile(out, decReader); err != nil {
		return nil, DefaultEncrypt, err
	}
	if err := inFile.Close(); err != nil {
		return nil, DefaultEncrypt, err
	}

	var sender *Key
	if key != nil {
		s, err := s.findSender(ctx, key.ID())
		if err != nil {
			return nil, DefaultEncrypt, err
		}
		sender = s
	}

	return sender, mode, nil
}

func writeFile(out string, reader io.Reader) error {
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

	if _, err := writer.ReadFrom(reader); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
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
