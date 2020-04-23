package service

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncryptDecryptCommand(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))

	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()

	client, closeClFn := newTestRPCClient(t, service, env)
	defer closeClFn()

	testAuthSetup(t, service)
	testImportKey(t, service, alice)
	testImportID(t, service, bob.ID())

	inPath := writeTestFile(t)
	outPath := inPath + ".enc"
	defer os.Remove(inPath)
	defer os.Remove(outPath)

	cmd := append(os.Args[0:1], "-app", env.appName) // "-log-level=debug"

	var clientErr error
	errorFn := func(err error) {
		clientErr = err
	}

	build := Build{Version: VersionDev}

	// Default
	argsEncrypt := append(cmd, "encrypt", "-r", alice.ID().String(), "-r", bob.ID().String(), "-in", inPath)
	runClient(build, argsEncrypt, client, errorFn)
	require.NoError(t, clientErr)
	os.Remove(inPath)

	argsDecrypt := append(cmd, "decrypt", "-in", outPath)
	runClient(build, argsDecrypt, client, errorFn)
	require.NoError(t, clientErr)

	in, err := ioutil.ReadFile(inPath)
	require.NoError(t, err)
	require.Equal(t, string(in), "test message")

	// Armored
	argsEncrypt = append(cmd, "encrypt", "-r", alice.ID().String(), "-r", bob.ID().String(), "-a", "-in", inPath)
	runClient(build, argsEncrypt, client, errorFn)
	require.NoError(t, clientErr)
	os.Remove(inPath)
	out, err := ioutil.ReadFile(inPath + ".enc")
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(out), "BEGIN SALTPACK ENCRYPTED MESSAGE"))

	argsDecrypt = append(cmd, "decrypt", "-a", "-in", outPath)
	runClient(build, argsDecrypt, client, errorFn)
	require.NoError(t, clientErr)

	in, err = ioutil.ReadFile(inPath)
	require.NoError(t, err)
	require.Equal(t, string(in), "test message")

	// Not found
	argsEncrypt = append(cmd, "encrypt", "-r", alice.ID().String(), "-r", bob.ID().String(), "-in", inPath+".notfound")
	runClient(build, argsEncrypt, client, errorFn)
	// TODO: This error
	if runtime.GOOS == "windows" {
		require.EqualError(t, clientErr, fmt.Sprintf("rpc error: code = Unknown desc = open %s: The system cannot find the file specified.", inPath+".notfound"))
	} else {
		require.EqualError(t, clientErr, fmt.Sprintf("rpc error: code = Unknown desc = open %s: no such file or directory", inPath+".notfound"))
	}
	clientErr = nil
}
