package service

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSignVerifyCommand(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))

	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()

	client, closeClFn := newTestRPCClient(t, service, env)
	defer closeClFn()

	testAuthSetup(t, service)
	testImportKey(t, service, alice)

	var clientErr error
	errorFn := func(err error) {
		clientErr = err
	}

	build := Build{Version: VersionDev}

	// Default
	inPath := writeTestFile(t)
	outPath := inPath + ".sig"
	defer os.Remove(inPath)
	defer os.Remove(outPath)

	cmd := append(os.Args[0:1], "-app", env.appName) // "-log-level=debug"

	argsSign := append(cmd, "sign", "-s", alice.ID().String(), "-in", inPath)
	runClient(build, argsSign, client, errorFn)
	require.NoError(t, clientErr)
	require.FileExists(t, outPath)
	os.Remove(inPath)

	argsVerify := append(cmd, "verify", "-in", outPath)
	runClient(build, argsVerify, client, errorFn)
	require.NoError(t, clientErr)

	in, err := ioutil.ReadFile(inPath)
	require.NoError(t, err)
	require.Equal(t, string(in), "test message")

	// Armored
	inPath = writeTestFile(t)
	outPath = inPath + ".sig"
	defer os.Remove(inPath)
	defer os.Remove(outPath)

	argsSign = append(cmd, "sign", "-a", "-s", alice.ID().String(), "-in", inPath)
	runClient(build, argsSign, client, errorFn)
	require.NoError(t, clientErr)
	require.FileExists(t, outPath)
	os.Remove(inPath)

	out, err := ioutil.ReadFile(inPath + ".sig")
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(out), "BEGIN SALTPACK SIGNED MESSAGE."))

	argsVerify = append(cmd, "verify", "-a", "-in", outPath)
	runClient(build, argsVerify, client, errorFn)
	require.NoError(t, clientErr)

	in, err = ioutil.ReadFile(inPath)
	require.NoError(t, err)
	require.Equal(t, string(in), "test message")

	// Detached
	inPath = writeTestFile(t)
	sigPath := inPath + ".sig"
	defer os.Remove(inPath)
	defer os.Remove(sigPath)

	argsSign = append(cmd, "sign", "-d", "-s", alice.ID().String(), "-in", inPath)
	runClient(build, argsSign, client, errorFn)
	require.NoError(t, clientErr)
	require.FileExists(t, sigPath)

	argsVerify = append(cmd, "verify", "-x", sigPath, "-in", inPath)
	runClient(build, argsVerify, client, errorFn)
	require.NoError(t, clientErr)

	in, err = ioutil.ReadFile(inPath)
	require.NoError(t, err)
	require.Equal(t, string(in), "test message")

	// Armored/Detached
	inPath = writeTestFile(t)
	sigPath = inPath + ".sig"
	defer os.Remove(inPath)
	defer os.Remove(sigPath)

	argsSign = append(cmd, "sign", "-d", "-a", "-s", alice.ID().String(), "-in", inPath)
	runClient(build, argsSign, client, errorFn)
	require.NoError(t, clientErr)
	require.FileExists(t, sigPath)
	out, err = ioutil.ReadFile(inPath + ".sig")
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(out), "BEGIN SALTPACK DETACHED SIGNATURE."))

	argsVerify = append(cmd, "verify", "-a", "-x", sigPath, "-in", inPath)
	runClient(build, argsVerify, client, errorFn)
	require.NoError(t, clientErr)

	in, err = ioutil.ReadFile(inPath)
	require.NoError(t, err)
	require.Equal(t, string(in), "test message")
}
