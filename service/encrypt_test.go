package service

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/keys-pub/keys/saltpack"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	testEncryptDecrypt(t, EncryptV2)
	testEncryptDecrypt(t, SigncryptV1)
}

func testEncryptDecrypt(t *testing.T, mode EncryptMode) {
	// SetLogger(newLog(DebugLevel))
	// saltpack.SetLogger(newLog(DebugLevel))
	// client.SetLogger(newLog(DebugLevel))
	env := newTestEnv(t)

	aliceService, aliceCloseFn := newTestService(t, env)
	defer aliceCloseFn()
	testAuthSetup(t, aliceService, alice)

	bobService, bobCloseFn := newTestService(t, env)
	defer bobCloseFn()
	testAuthSetup(t, bobService, bob)

	message := "Hey bob"

	// Encrypt
	encryptResp, err := aliceService.Encrypt(context.TODO(), &EncryptRequest{
		Data:       []byte(message),
		Sender:     alice.ID().String(),
		Recipients: []string{bob.ID().String()},
		Mode:       mode,
	})
	require.NoError(t, err)
	require.NotEmpty(t, encryptResp.Data)

	// Decrypt
	decryptResp, err := bobService.Decrypt(context.TODO(), &DecryptRequest{
		Data: encryptResp.Data,
		Mode: mode,
	})
	require.NoError(t, err)
	require.Equal(t, message, string(decryptResp.Data))

	// Alice try to decrypt her own message
	// TODO: Include alice by default?
	_, err = aliceService.Decrypt(context.TODO(), &DecryptRequest{
		Data: encryptResp.Data,
		Mode: mode,
	})
	require.EqualError(t, err, "no decryption key found for message")

	_, err = aliceService.Encrypt(context.TODO(), &EncryptRequest{
		Data:   []byte(message),
		Sender: alice.ID().String(),
		Mode:   mode,
	})
	require.EqualError(t, err, "no recipients specified")
}

func TestEncryptAnonymous(t *testing.T) {
	env := newTestEnv(t)

	aliceService, aliceCloseFn := newTestService(t, env)
	defer aliceCloseFn()
	testAuthSetup(t, aliceService, alice)

	bobService, bobCloseFn := newTestService(t, env)
	defer bobCloseFn()
	testAuthSetup(t, bobService, bob)

	message := "Hey bob"

	// Encrypt
	encryptResp, err := aliceService.Encrypt(context.TODO(), &EncryptRequest{
		Data:       []byte(message),
		Sender:     "",
		Recipients: []string{bob.ID().String()},
	})
	require.NoError(t, err)
	require.NotEmpty(t, encryptResp.Data)

	// Decrypt
	decryptResp, err := bobService.Decrypt(context.TODO(), &DecryptRequest{
		Data: encryptResp.Data,
	})
	require.NoError(t, err)
	require.Equal(t, message, string(decryptResp.Data))

	// Encrypt
	_, err = aliceService.Encrypt(context.TODO(), &EncryptRequest{
		Data:       []byte(message),
		Sender:     "",
		Recipients: []string{bob.ID().String()},
		Mode:       SigncryptV1,
	})
	require.EqualError(t, err, "no sender specified: sender is required for signcrypt mode")
}

func TestEncryptStream(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	testAuthSetup(t, service, alice)
	testImportKey(t, service, bob)

	testEncryptStream(t, service, bytes.Repeat([]byte{0x31}, 5), alice.ID().String(), []string{bob.ID().String()})
	testEncryptStream(t, service, bytes.Repeat([]byte{0x31}, (1024*1024)+5), alice.ID().String(), []string{bob.ID().String()})
	// TODO: Test timeout if data stops streaming
}

func testEncryptStream(t *testing.T, service *service, plaintext []byte, sender string, recipients []string) {
	client, clientCloseFn := newTestRPCClient(t, service)
	defer clientCloseFn()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	// Seal stream
	streamClient, streamErr := client.ProtoClient().EncryptStream(ctx)
	require.NoError(t, streamErr)

	chunkSize := 1024 * 1024
	go func() {
		done := false
		err := streamClient.Send(&EncryptStreamInput{
			Recipients: recipients,
			Sender:     sender,
			Armored:    true,
			Mode:       SigncryptV1,
		})
		require.NoError(t, err)
		for chunk := 0; true; chunk++ {
			s, e := (chunk * chunkSize), ((chunk + 1) * chunkSize)
			if e > len(plaintext) {
				e = len(plaintext)
				done = true
			}
			logger.Debugf("(Test) Send chunk %d", len(plaintext[s:e]))
			err := streamClient.Send(&EncryptStreamInput{
				Data: plaintext[s:e],
			})
			require.NoError(t, err)
			if done {
				logger.Debugf("(Test) Send done")
				break
			}
		}
		logger.Debugf("(Test) Close send")
		closeErr := streamClient.CloseSend()
		require.NoError(t, closeErr)
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

	// Decrypt (from Saltpack)
	encrypted := buf.Bytes()
	sp := saltpack.NewSaltpack(service.ks)
	sp.SetArmored(true)
	out, signer, err := sp.SigncryptOpen(encrypted)
	require.NoError(t, err)
	if sender != "" {
		require.Equal(t, sender, signer.String())
	}
	require.Equal(t, plaintext, out)

	// Decrypt stream
	outClient, streamErr2 := client.ProtoClient().SigncryptOpenArmoredStream(ctx)
	require.NoError(t, streamErr2)

	go func() {
		done := false
		for chunk := 0; ; chunk++ {
			s, e := (chunk * chunkSize), ((chunk + 1) * chunkSize)
			if e > len(encrypted) {
				e = len(encrypted)
				done = true
			}
			err := outClient.Send(&DecryptStreamInput{
				Data: encrypted[s:e],
			})
			require.NoError(t, err)
			if done {
				break
			}
		}
		closeErr := outClient.CloseSend()
		require.NoError(t, closeErr)
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
