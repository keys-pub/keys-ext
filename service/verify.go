package service

import (
	"context"
	"io"
	"strings"

	"github.com/keys-pub/keys/saltpack"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Verify (RPC) ...
func (s *service) Verify(ctx context.Context, req *VerifyRequest) (*VerifyResponse, error) {
	verified, kid, err := saltpack.Verify(req.Data)
	if err != nil {
		return nil, err
	}

	var signer *Key
	if kid != "" {
		s, err := s.verifyKey(ctx, kid)
		if err != nil {
			return nil, err
		}
		signer = s
	}

	return &VerifyResponse{Data: verified, Signer: signer}, nil
}

// VerifyDetached (RPC) ...
func (s *service) VerifyDetached(ctx context.Context, req *VerifyDetachedRequest) (*VerifyDetachedResponse, error) {
	kid, err := saltpack.VerifyDetached(req.Sig, req.Data)
	if err != nil {
		return nil, err
	}

	var signer *Key
	if kid != "" {
		s, err := s.verifyKey(ctx, kid)
		if err != nil {
			return nil, err
		}
		signer = s
	}

	return &VerifyDetachedResponse{Signer: signer}, nil
}

// VerifyFile (RPC) ...
func (s *service) VerifyFile(srv Keys_VerifyFileServer) error {
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
		if strings.HasSuffix(in, ".signed") {
			out = strings.TrimSuffix(in, ".signed")
		} else {
			out = in + ".verified"
		}
	}
	outExists, err := pathExists(out)
	if err != nil {
		return err
	}
	if outExists {
		return errors.Errorf("file already exists %s", out)
	}

	signer, err := s.verifyWriteInOut(srv.Context(), in, out)
	if err != nil {
		return errors.Wrapf(err, "failed to verify")
	}

	if err := srv.Send(&VerifyFileOutput{
		Signer: signer,
		Out:    out,
	}); err != nil {
		return err
	}

	// if err := srv.SendAndClose(&VerifyFileOutput{
	// 	Signer: signer,
	// 	Out:    out,
	// }); err != nil {
	// 	return err
	// }

	return nil
}

// VerifyDetachedFile (RPC) ...
func (s *service) VerifyDetachedFile(srv Keys_VerifyDetachedFileServer) error {
	req, err := srv.Recv()
	if err != nil {
		return err
	}
	in := req.In
	if in == "" {
		return errors.Errorf("in not specified")
	}

	signer, err := s.verifyDetachedIn(srv.Context(), req.Sig, in)
	if err != nil {
		return err
	}

	resp := &VerifyDetachedResponse{
		Signer: signer,
	}
	return srv.SendAndClose(resp)
}

// VerifyStream (RPC) ...
func (s *service) VerifyStream(srv Keys_VerifyStreamServer) error {
	return s.verifyStream(srv, false)
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

	streamReader, kid, err := saltpack.NewVerifyStream(reader)
	if err != nil {
		return errors.Wrapf(err, "faild to verify stream")
	}

	var signer *Key
	if kid != "" {
		s, err := s.verifyKey(srv.Context(), kid)
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
	if err := s.readFromStream(srv.Context(), streamReader, signer, sendFn); err != nil {
		return errors.Wrapf(err, "failed to read from verify stream")
	}
	return nil
}

// VerifyDetachedStream (RPC) ...
func (s *service) VerifyDetachedStream(srv Keys_VerifyDetachedStreamServer) error {
	ctx := srv.Context()

	first, err := srv.Recv()
	if err != nil {
		return err
	}

	recvFn := func() ([]byte, error) {
		req, recvErr := srv.Recv()
		if recvErr != nil {
			return nil, recvErr
		}
		return req.Data, nil
	}

	reader := newStreamReader(srv.Context(), recvFn)
	if err := reader.write(first.Data); err != nil {
		return err
	}
	kid, err := saltpack.VerifyDetachedReader(first.Sig, reader)
	if err != nil {
		return err
	}
	var signer *Key
	if kid != "" {
		s, err := s.verifyKey(ctx, kid)
		if err != nil {
			return err
		}
		signer = s
	}
	resp := &VerifyDetachedResponse{
		Signer: signer,
	}
	return srv.SendAndClose(resp)
}

// VerifyStreamClient ...
type VerifyStreamClient interface {
	Send(*VerifyInput) error
	Recv() (*VerifyOutput, error)
	grpc.ClientStream
}

// NewVerifyStreamClient ...
func NewVerifyStreamClient(ctx context.Context, cl KeysClient) (VerifyStreamClient, error) {
	return cl.VerifyStream(ctx)
}

func (s *service) readFromStream(ctx context.Context, streamReader io.Reader, signer *Key, sendFn func(b []byte, signer *Key) error) error {
	buf := make([]byte, 1024*1024)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, err := streamReader.Read(buf)
		if n != 0 {
			if err := sendFn(buf[:n], signer); err != nil {
				return err
			}
		}
		if err != nil {
			if err == io.EOF {
				// Send empty bytes denotes EOF
				if err := sendFn([]byte{}, signer); err != nil {
					return err
				}
				break
			}
			return err
		}
	}

	return nil
}

func (s *service) verifyWriteInOut(ctx context.Context, in string, out string) (*Key, error) {
	kid, err := saltpack.VerifyFile(in, out)
	if err != nil {
		return nil, err
	}

	var signer *Key
	if kid != "" {
		s, err := s.verifyKey(ctx, kid)
		if err != nil {
			return nil, err
		}
		signer = s
	}

	return signer, nil
}

func (s *service) verifyDetachedIn(ctx context.Context, sig []byte, in string) (*Key, error) {
	logger.Infof("Verify (detached) %s", in)

	kid, err := saltpack.VerifyFileDetached(sig, in)
	if err != nil {
		return nil, err
	}

	var signer *Key
	if kid != "" {
		s, err := s.verifyKey(ctx, kid)
		if err != nil {
			return nil, err
		}
		signer = s
	}

	return signer, nil
}
