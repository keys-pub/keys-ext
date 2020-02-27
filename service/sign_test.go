package service

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/keys-pub/keys/saltpack"
	"github.com/stretchr/testify/require"
)

func TestSignVerify(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	testAuthSetup(t, service)
	testImportKey(t, service, alice)

	message := "I'm alice"
	signResp, err := service.Sign(context.TODO(), &SignRequest{Data: []byte(message), Signer: alice.ID().String()})
	require.NoError(t, err)
	require.NotEmpty(t, signResp.Data)
	require.Equal(t, alice.ID().String(), signResp.KID)

	verifyResp, err := service.Verify(context.TODO(), &VerifyRequest{Data: signResp.Data})
	require.NoError(t, err)
	require.Equal(t, message, string(verifyResp.Data))
	require.Equal(t, alice.ID().String(), verifyResp.Signer.ID)

	testUserSetupGithub(t, env, service, alice, "alice")

	signResp, err = service.Sign(context.TODO(), &SignRequest{Data: []byte(message), Signer: "alice@github"})
	require.NoError(t, err)
	require.NotEmpty(t, signResp.Data)
	require.Equal(t, alice.ID().String(), signResp.KID)

	verifyResp, err = service.Verify(context.TODO(), &VerifyRequest{Data: signResp.Data})
	require.NoError(t, err)
	require.Equal(t, message, string(verifyResp.Data))
	require.Equal(t, alice.ID().String(), verifyResp.Signer.ID)
}

func TestSignStream(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	testAuthSetup(t, service)
	testImportKey(t, service, alice)

	testSignStream(t, service, bytes.Repeat([]byte{0x31}, 5), alice.ID().String())
	testSignStream(t, service, bytes.Repeat([]byte{0x31}, (1024*1024)+5), alice.ID().String())
	// TODO: Test timeout if data stops streaming
}

func testSignStream(t *testing.T, service *service, plaintext []byte, signer string) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	cl, clientCloseFn := newTestRPCClient(t, service)
	defer clientCloseFn()

	streamClient, streamErr := cl.ProtoClient().SignStream(ctx)
	require.NoError(t, streamErr)

	chunkSize := 1024 * 1024
	go func() {
		done := false
		err := streamClient.Send(&SignInput{
			Signer:  signer,
			Armored: true,
		})
		require.NoError(t, err)
		for chunk := 0; true; chunk++ {
			s, e := (chunk * chunkSize), ((chunk + 1) * chunkSize)
			if e > len(plaintext) {
				e = len(plaintext)
				done = true
			}
			logger.Debugf("(Test) Send chunk %d", len(plaintext[s:e]))
			err := streamClient.Send(&SignInput{
				Data: plaintext[s:e],
			})
			require.NoError(t, err)
			if done {
				logger.Debugf("(Test) Send done")
				break
			}
		}
		logger.Debugf("(Test) Close send")
		closeErr2 := streamClient.CloseSend()
		require.NoError(t, closeErr2)
	}()

	var buf bytes.Buffer
	for {
		resp, recvErr := streamClient.Recv()
		if recvErr != nil {
			if recvErr == io.EOF {
				logger.Debugf("(Test) Recv EOF, done")
				break
			}
			require.NoError(t, recvErr)
		}
		logger.Infof("(Test) Recv %d", len(resp.Data))
		_, writeErr := buf.Write(resp.Data)
		require.NoError(t, writeErr)
	}

	// Verify (from Saltpack)
	data := buf.Bytes()
	sp := saltpack.NewSaltpack(service.ks)
	out, sout, err := sp.VerifyArmored(string(data))
	require.NoError(t, err)
	if signer != "" {
		require.Equal(t, sout.String(), signer)
	}
	require.Equal(t, plaintext, out)

	// Verify stream
	outClient, streamErr := cl.ProtoClient().VerifyArmoredStream(ctx)
	require.NoError(t, streamErr)

	go func() {
		done := false
		for chunk := 0; ; chunk++ {
			s, e := (chunk * chunkSize), ((chunk + 1) * chunkSize)
			if e > len(data) {
				e = len(data)
				done = true
			}
			err := outClient.Send(&VerifyInput{
				Data: data[s:e],
			})
			require.NoError(t, err)
			if done {
				break
			}
		}
		closeErr2 := outClient.CloseSend()
		require.NoError(t, closeErr2)
	}()

	var bufOut bytes.Buffer
	for {
		resp, recvErr := outClient.Recv()
		require.NoError(t, recvErr)
		if len(resp.Data) == 0 {
			break
		}
		_, writeErr := bufOut.Write(resp.Data)
		require.NoError(t, writeErr)
	}

	require.Equal(t, plaintext, bufOut.Bytes())
}

func TestSignVerifyFile(t *testing.T) {
	env := newTestEnv(t)

	aliceService, aliceCloseFn := newTestService(t, env)
	defer aliceCloseFn()
	testAuthSetup(t, aliceService)
	testImportKey(t, aliceService, alice)

	bobService, bobCloseFn := newTestService(t, env)
	defer bobCloseFn()
	testAuthSetup(t, bobService)
	testImportKey(t, bobService, bob)

	b := []byte("test message")
	inPath := filepath.Join(os.TempDir(), "test.txt")
	outPath := inPath + ".sig"
	verifiedPath := inPath + ".ver"

	defer os.Remove(inPath)
	defer os.Remove(outPath)
	defer os.Remove(verifiedPath)

	writeErr := ioutil.WriteFile(inPath, b, 0644)
	require.NoError(t, writeErr)

	aliceClient, aliceClientCloseFn := newTestRPCClient(t, aliceService)
	defer aliceClientCloseFn()

	err := signFile(aliceClient, alice.ID().String(), true, false, inPath, outPath)
	require.NoError(t, err)

	bobClient, bobClientCloseFn := newTestRPCClient(t, bobService)
	defer bobClientCloseFn()

	ver, err := verifyFile(bobClient, true, outPath, verifiedPath)
	require.NoError(t, err)
	require.NotNil(t, ver.Signer)
	require.Equal(t, alice.ID().String(), ver.Signer.ID)

	bout, err := ioutil.ReadFile(verifiedPath)
	require.NoError(t, err)
	require.Equal(t, b, bout)
	os.Remove(verifiedPath)

	ver, err = verifyFile(bobClient, true, outPath, "")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(os.TempDir(), "test-1.txt"), ver.Out)
	os.Remove(ver.Out)
}
