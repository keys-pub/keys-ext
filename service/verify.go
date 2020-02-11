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

type verify struct {
	armored bool
	sp      *saltpack.Saltpack
}

func (s *service) newVerify(armored bool) *verify {
	sp := saltpack.NewSaltpack(s.ks)
	sp.SetArmored(armored)
	return &verify{armored: armored, sp: sp}
}

// Verify (RPC) ...
func (s *service) Verify(ctx context.Context, req *VerifyRequest) (*VerifyResponse, error) {
	ver := s.newVerify(req.Armored)

	verified, kid, err := ver.sp.Verify(req.Data)
	if err != nil {
		return nil, err
	}

	var signer *Key
	if kid != "" {
		s, err := s.loadKey(ctx, kid)
		if err != nil {
			return nil, err
		}
		signer = s
	}

	return &VerifyResponse{Data: verified, Signer: signer}, nil
}

// VerifyFile (RPC) ...
func (s *service) VerifyFile(srv Keys_VerifyFileServer) error {
	req, err := srv.Recv()
	if err != nil {
		return err
	}
	if req.In == "" {
		return errors.Errorf("in not specified")
	}
	out := req.Out
	if out == "" {
		if strings.HasSuffix(req.In, ".sig") {
			out = strings.TrimSuffix(req.In, ".sig")
		}
		out, err = nextPath(out)
		if err != nil {
			return err
		}
	}

	ver := s.newVerify(req.Armored)

	signer, err := s.verifyWriteInOut(srv.Context(), req.In, out, ver)
	if err != nil {
		return err
	}

	if err := srv.Send(&VerifyFileOutput{
		Signer: signer,
		Out:    out,
	}); err != nil {
		return err
	}

	return nil
}

// VerifyStream (RPC) ...
func (s *service) VerifyStream(srv Keys_VerifyStreamServer) error {
	return s.verifyStream(srv, false)
}

// VerifyArmoredStream (RPC) ...
func (s *service) VerifyArmoredStream(srv Keys_VerifyArmoredStreamServer) error {
	return s.verifyStream(srv, true)
}

type verifyStreamServer interface {
	Send(*VerifyOutput) error
	Recv() (*VerifyInput, error)
	grpc.ServerStream
}

func (s *service) verifyStream(srv verifyStreamServer, armored bool) error {
	recvFn := func() ([]byte, error) {
		req, recvErr := srv.Recv()
		if recvErr != nil {
			return nil, recvErr
		}
		return req.Data, nil
	}

	reader := newStreamReader(srv.Context(), recvFn)

	ver := s.newVerify(armored)
	streamReader, kid, err := s.verifyReader(srv.Context(), reader, ver)
	if err != nil {
		return err
	}

	var signer *Key
	if kid != "" {
		s, err := s.loadKey(srv.Context(), kid)
		if err != nil {
			return err
		}
		signer = s
	}
	sendFn := func(b []byte, signer *Key) error {
		resp := VerifyOutput{
			Data:   b,
			Signer: signer,
		}
		return srv.Send(&resp)
	}
	return s.readFromStream(srv.Context(), streamReader, signer, sendFn)
}

// VerifyStreamClient ...
type VerifyStreamClient interface {
	Send(*VerifyInput) error
	Recv() (*VerifyOutput, error)
	grpc.ClientStream
}

// NewVerifyStreamClient ...
func NewVerifyStreamClient(ctx context.Context, cl KeysClient, armored bool) (VerifyStreamClient, error) {
	if armored {
		return cl.VerifyArmoredStream(ctx)
	}
	return cl.VerifyStream(ctx)
}

func (s *service) readFromStream(ctx context.Context, streamReader io.Reader, signer *Key, sendFn func(b []byte, signer *Key) error) error {
	buf := make([]byte, 1024*1024)
	sendSigner := signer
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, err := streamReader.Read(buf)
		if n != 0 {
			if err := sendFn(buf[:n], sendSigner); err != nil {
				return err
			}
			// Only send signer on first send
			sendSigner = nil
		}
		if err != nil {
			if err == io.EOF {
				// Send empty bytes denotes EOF
				if err := sendFn([]byte{}, nil); err != nil {
					return err
				}
				break
			}
			return err
		}

	}

	return nil
}

func (s *service) verifyReader(ctx context.Context, reader io.Reader, ver *verify) (io.Reader, keys.ID, error) {
	return ver.sp.NewVerifyStream(reader)
}

func (s *service) verifyWriteInOut(ctx context.Context, in string, out string, ver *verify) (*Key, error) {
	inFile, err := os.Open(in)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(inFile)

	verifyReader, kid, err := s.verifyReader(ctx, reader, ver)
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

	if _, err := writer.ReadFrom(verifyReader); err != nil {
		return nil, err
	}
	writer.Flush()

	if err := os.Rename(outTmp, out); err != nil {
		return nil, err
	}

	var signer *Key
	if kid != "" {
		s, err := s.loadKey(ctx, kid)
		if err != nil {
			return nil, err
		}
		signer = s
	}

	return signer, nil
}
