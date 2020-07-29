package service

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSignVerifyCommand(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))

	env := newTestEnv(t)
	appName := "KeysTest-" + randName()
	service, closeFn := newTestService(t, env, appName)
	var clientOut bytes.Buffer
	client, closeClFn := newTestRPCClient(t, service, env, appName, &clientOut)
	defer closeClFn()
	defer closeFn()

	testAuthSetup(t, service)
	testImportKey(t, service, alice)

	var clientErr error
	errorFn := func(err error) {
		clientErr = err
	}

	build := Build{Version: VersionDev}

	// Default (armored, detached) (file)
	inPath := writeTestFile(t)
	sigPath := inPath + ".sig"
	defer os.Remove(inPath)
	defer os.Remove(sigPath)

	cmd := append(os.Args[0:1], "-app", appName) // , "-log-level=debug")

	// Default: Armored, detached (file)
	argsSign := append(cmd, "sign", "-s", alice.ID().String(), "-in", inPath)
	runClient(build, argsSign, client, errorFn)
	require.NoError(t, clientErr)
	require.FileExists(t, sigPath)
	sig, err := ioutil.ReadFile(sigPath)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(sig), "BEGIN SALTPACK DETACHED SIGNATURE."))
	require.Equal(t, fmt.Sprintf("out: %s\n", sigPath), clientOut.String())
	clientOut.Reset()

	argsVerify := append(cmd, "verify", "-s", alice.ID().String(), "-in", inPath, "-x", inPath+".sig")
	runClient(build, argsVerify, client, errorFn)
	require.NoError(t, clientErr)
	require.Equal(t, "", clientOut.String())
	clientOut.Reset()

	// Binary, detached (file)
	inPath = writeTestFile(t)
	sigPath = inPath + ".sig"
	defer os.Remove(inPath)
	defer os.Remove(sigPath)

	argsSign = append(cmd, "sign", "-binary", "-s", alice.ID().String(), "-in", inPath)
	runClient(build, argsSign, client, errorFn)
	require.NoError(t, clientErr)
	require.FileExists(t, sigPath)
	require.Equal(t, fmt.Sprintf("out: %s\n", sigPath), clientOut.String())
	clientOut.Reset()

	argsVerify = append(cmd, "verify", "-s", alice.ID().String(), "-in", inPath, inPath, "-x", inPath+".sig")
	runClient(build, argsVerify, client, errorFn)
	require.NoError(t, clientErr)
	require.Equal(t, "", clientOut.String())
	clientOut.Reset()

	// Armored, attached (file)
	inPath = writeTestFile(t)
	outPath := inPath + ".signed"
	defer os.Remove(inPath)
	defer os.Remove(sigPath)

	argsSign = append(cmd, "sign", "-armor", "-attached", "-s", alice.ID().String(), "-in", inPath)
	runClient(build, argsSign, client, errorFn)
	require.NoError(t, clientErr)
	require.FileExists(t, outPath)
	signed, err := ioutil.ReadFile(outPath)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(signed), "BEGIN SALTPACK SIGNED MESSAGE."))
	os.Remove(inPath)
	require.Equal(t, fmt.Sprintf("out: %s\n", outPath), clientOut.String())
	clientOut.Reset()

	argsVerify = append(cmd, "verify", "-s", alice.ID().String(), "-in", outPath)
	runClient(build, argsVerify, client, errorFn)
	require.NoError(t, clientErr)
	require.Equal(t, "", clientOut.String())
	clientOut.Reset()

	in, err := ioutil.ReadFile(inPath)
	require.NoError(t, err)
	require.Equal(t, string(in), "test message")

}
