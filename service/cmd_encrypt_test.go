package service

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncryptDecryptCommand(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))

	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()

	client, closeClFn := newTestRPCClient(t, service)
	defer closeClFn()

	testAuthSetup(t, service)
	testImportKey(t, service, alice)
	testImportID(t, service, bob.ID())

	inPath := filepath.Join(os.TempDir(), "test.txt")
	outPath := inPath + ".enc"
	defer os.Remove(inPath)
	defer os.Remove(outPath)

	var clientErr error
	errorFn := func(err error) {
		clientErr = err
	}

	argsEncrypt := append(os.Args[0:1], "-test", "encrypt", "-r", alice.ID().String(), "-r", bob.ID().String(), "-in", inPath, "-out", outPath)
	runClient(Build{Version: "1.2.3"}, argsEncrypt, client, errorFn)
	require.EqualError(t, clientErr, fmt.Sprintf("rpc error: code = Unknown desc = open %s: no such file or directory", inPath))
	clientErr = nil

	writeErr := ioutil.WriteFile(inPath, []byte("test message"), 0644)
	require.NoError(t, writeErr)

	// Default
	runClient(Build{Version: "1.2.3"}, argsEncrypt, client, errorFn)
	require.NoError(t, clientErr)

	outPath2 := inPath + ".out"
	argsDecrypt := append(os.Args[0:1], "-test", "decrypt", "-in", outPath, "-out", outPath2)
	runClient(Build{Version: "1.2.3"}, argsDecrypt, client, errorFn)
	require.NoError(t, clientErr)

	out, err := ioutil.ReadFile(outPath2)
	require.NoError(t, err)

	require.Equal(t, string(out), "test message")

	// Armored
	argsEncrypt = append(os.Args[0:1], "-test", "encrypt", "-r", alice.ID().String(), "-r", bob.ID().String(), "-a", "-in", inPath, "-out", outPath)
	runClient(Build{Version: "1.2.3"}, argsEncrypt, client, errorFn)
	require.NoError(t, clientErr)

	outPath2 = inPath + ".out"
	argsDecrypt = append(os.Args[0:1], "-test", "decrypt", "-a", "-in", outPath, "-out", outPath2)
	runClient(Build{Version: "1.2.3"}, argsDecrypt, client, errorFn)
	require.NoError(t, clientErr)

	out, err = ioutil.ReadFile(outPath2)
	require.NoError(t, err)

	require.Equal(t, string(out), "test message")
}
