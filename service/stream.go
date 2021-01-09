package service

import (
	"bytes"
	"context"

	"github.com/pkg/errors"
)

type streamReader struct {
	ctx    context.Context
	recvFn func() ([]byte, error)
	buf    bytes.Buffer
}

func newStreamReader(ctx context.Context, recvFn func() ([]byte, error)) *streamReader {
	return &streamReader{
		ctx:    ctx,
		recvFn: recvFn,
		buf:    bytes.Buffer{},
	}
}

func (r *streamReader) write(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	n, err := r.buf.Write(b)
	if err != nil {
		return err
	}
	if n != len(b) {
		return errors.Errorf("failed to write all bytes %d != %d", n, len(b))
	}
	return nil
}

func (r *streamReader) recv() error {
	b, recvErr := r.recvFn()
	if recvErr != nil {
		return recvErr
	}
	return r.write(b)
}

func (r *streamReader) Read(p []byte) (n int, err error) {
	// Only recv when buffer is empty and more bytes are requested, otherwise
	// recv will block forever.
	if r.buf.Len() == 0 {
		if err := r.recv(); err != nil {
			return 0, err
		}
	}
	return r.buf.Read(p)
}
