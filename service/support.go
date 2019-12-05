package service

import (
	"bufio"
	"encoding/json"
	"io"
	"os/exec"
	"runtime"

	"github.com/pkg/errors"
)

func checkSupportedOS() error {
	switch rt := runtime.GOOS; rt {
	case "darwin", "windows", "linux":
		return nil
	default:
		return errors.Errorf("%s is not currently supported", rt)
	}
}

func checkCodesigned() error {
	exe, exeErr := ExecutablePath()
	if exeErr != nil {
		return exeErr
	}
	cmd := exec.Command("/usr/bin/codesign", "-v", exe)
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "%s is not codesigned", exe)
	}
	return nil
}

func readFrom(reader io.Reader, chunkSize int, processFn func([]byte) error) (int64, error) {
	numBytes := int64(0)
	r := bufio.NewReader(reader)
	buf := make([]byte, 0, chunkSize)
	for {
		n, err := r.Read(buf[:cap(buf)])
		buf = buf[:n]
		if n > 0 {
			if err := processFn(buf); err != nil {
				return numBytes, err
			}
			numBytes += int64(len(buf))
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return numBytes, err
		}
	}
	if err := processFn([]byte{}); err != nil {
		return numBytes, err
	}
	return numBytes, nil
}

func mustJSONMarshal(i interface{}) string {
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	return string(b)
}
