package service

import (
	"bufio"
	"io"
	"os"
)

func readerFromArgs(in string) (io.Reader, error) {
	if in != "" {
		file, err := os.Open(in)
		if err != nil {
			return nil, err
		}
		return bufio.NewReader(file), nil
	}
	return bufio.NewReader(os.Stdin), nil
}

func writerFromArgs(out string) (io.Writer, error) {
	if out != "" {
		file, createErr := os.Create(out)
		if createErr != nil {
			return nil, createErr
		}
		return file, nil
	}
	return os.Stdout, nil
}
