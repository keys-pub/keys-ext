package service

import (
	"bufio"
	"context"
	"io"
	"os"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/saltpack"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type decrypt struct {
	mode EncryptMode
	sp   *saltpack.Saltpack
}

func (s *service) newDecrypt(ctx context.Context, armored bool, mode EncryptMode) *decrypt {
	if mode == DefaultEncryptMode {
		mode = EncryptV2
	}

	sp := saltpack.NewSaltpack(s.ks)
	sp.SetArmored(armored)
	return &decrypt{
		mode: mode,
		sp:   sp,
	}
}

func (s *service) decryptSigner(ctx context.Context, kid keys.ID, mode EncryptMode) (*Key, error) {
	if kid == "" {
		return nil, nil
	}
	// If EncryptV2 check the kid.
	if mode == EncryptV2 {
		k, err := s.checkSignerID(kid)
		if err != nil {
			return nil, err
		}
		kid = k
	}
	return s.loadKey(ctx, kid)
}

// Decrypt (RPC) data.
func (s *service) Decrypt(ctx context.Context, req *DecryptRequest) (*DecryptResponse, error) {
	dec := s.newDecrypt(ctx, req.Armored, req.Mode)

	var decrypted []byte
	var kid keys.ID
	var err error
	switch dec.mode {
	case EncryptV2:
		decrypted, kid, err = dec.sp.Decrypt(req.Data)
	case SigncryptV1:
		decrypted, kid, err = dec.sp.SigncryptOpen(req.Data)
	default:
		return nil, errors.Errorf("unsupported mode %s", dec.mode)
	}

	if err != nil {
		if err.Error() == "failed to read header bytes" {
			return nil, errors.Errorf("invalid data")
		}
		return nil, err
	}

	signer, err := s.decryptSigner(ctx, kid, dec.mode)
	if err != nil {
		return nil, err
	}

	return &DecryptResponse{
		Data:   decrypted,
		Signer: signer,
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
	if req.Out == "" {
		return errors.Errorf("out not specified")
	}

	dec := s.newDecrypt(srv.Context(), req.Armored, req.Mode)

	signer, err := s.decryptWriteInOut(srv.Context(), req.In, req.Out, dec)
	if err != nil {
		return err
	}

	if err := srv.Send(&DecryptFileOutput{
		Signer: signer,
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
	Send(*DecryptOutput) error
	Recv() (*DecryptInput, error)
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

	dec := s.newDecrypt(srv.Context(), armored, mode)

	reader := newStreamReader(srv.Context(), recvFn)

	streamReader, kid, err := s.decryptReader(srv.Context(), reader, dec)
	if err != nil {
		return err
	}

	signer, err := s.decryptSigner(srv.Context(), kid, dec.mode)
	if err != nil {
		return err
	}

	sendFn := func(b []byte, signer *Key) error {
		resp := DecryptOutput{
			Data:   b,
			Signer: signer,
		}
		return srv.Send(&resp)
	}
	return s.readFromStream(srv.Context(), streamReader, signer, sendFn)
}

func (s *service) decryptReader(ctx context.Context, reader io.Reader, dec *decrypt) (io.Reader, keys.ID, error) {
	switch dec.mode {
	case EncryptV2:
		return dec.sp.NewDecryptStream(reader)
	case SigncryptV1:
		return dec.sp.NewSigncryptOpenStream(reader)
	default:
		return nil, "", errors.Errorf("unsupported mode %s", dec.mode)
	}
}

func (s *service) decryptWriteInOut(ctx context.Context, in string, out string, dec *decrypt) (*Key, error) {
	inFile, err := os.Open(in)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(inFile)

	decReader, kid, err := s.decryptReader(ctx, reader, dec)
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
	writer.Flush()

	if err := os.Rename(outTmp, out); err != nil {
		return nil, err
	}

	signer, err := s.decryptSigner(ctx, kid, dec.mode)
	if err != nil {
		return nil, err
	}

	return signer, nil
}
