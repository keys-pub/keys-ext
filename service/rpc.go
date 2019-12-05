package service

import (
	"bytes"
	"context"
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

func (r *streamReader) recv() error {
	b, recvErr := r.recvFn()
	if recvErr != nil {
		return recvErr
	}
	_, writeErr := r.buf.Write(b)
	if writeErr != nil {
		return writeErr
	}
	return nil
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
