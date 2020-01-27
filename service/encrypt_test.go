package service

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/keys-pub/keys/saltpack"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/status"
)

func TestEncryptDecrypt(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// keys.SetLogger(NewLogger(DebugLevel))
	// saltpack.SetLogger(NewLogger(DebugLevel))
	// client.SetLogger(newLog(DebugLevel))
	env := newTestEnv(t)

	aliceService, aliceCloseFn := newTestService(t, env)
	defer aliceCloseFn()
	testAuthSetup(t, aliceService)
	testImportKey(t, aliceService, alice)

	bobService, bobCloseFn := newTestService(t, env)
	defer bobCloseFn()
	testAuthSetup(t, bobService)
	testImportKey(t, bobService, bob)

	// Save alice's sign public key to bob's keystore, check for box key conversion.
	err := bobService.ks.SaveSignPublicKey(alice.PublicKey())
	require.NoError(t, err)
	sk, err := bobService.ks.FindEd25519PublicKey(alice.Curve25519Key().PublicKey())
	require.NoError(t, err)
	require.NotNil(t, sk)
	require.Equal(t, alice.ID(), sk.ID())

	testEncryptDecrypt(t, aliceService, bobService, EncryptV2, true)
	testEncryptDecrypt(t, aliceService, bobService, EncryptV2, false)
	testEncryptDecrypt(t, aliceService, bobService, SigncryptV1, true)
	testEncryptDecrypt(t, aliceService, bobService, SigncryptV1, false)
}

func testEncryptDecrypt(t *testing.T, aliceService *service, bobService *service, mode EncryptMode, armored bool) {
	message := "Hey bob"

	// Encrypt
	encryptResp, err := aliceService.Encrypt(context.TODO(), &EncryptRequest{
		Data:       []byte(message),
		Sender:     alice.ID().String(),
		Recipients: []string{bob.ID().String()},
		Mode:       mode,
		Armored:    armored,
	})
	require.NoError(t, err)
	require.NotEmpty(t, encryptResp.Data)

	// Decrypt
	decryptResp, err := bobService.Decrypt(context.TODO(), &DecryptRequest{
		Data:    encryptResp.Data,
		Mode:    mode,
		Armored: armored,
	})
	require.NoError(t, err)
	require.Equal(t, message, string(decryptResp.Data))
	require.Equal(t, alice.ID().String(), decryptResp.Sender.ID)
}

func testEncryptDecryptErrors(t *testing.T, aliceService *service, bobService *service, mode EncryptMode, armored bool) {
	message := "Hey bob"

	encryptResp, err := aliceService.Encrypt(context.TODO(), &EncryptRequest{
		Data:       []byte(message),
		Sender:     alice.ID().String(),
		Recipients: []string{bob.ID().String()},
		Mode:       mode,
		Armored:    armored,
	})
	require.NoError(t, err)
	require.NotEmpty(t, encryptResp.Data)

	// Alice try to decrypt her own message
	// TODO: Include alice by default?
	_, err = aliceService.Decrypt(context.TODO(), &DecryptRequest{
		Data:    encryptResp.Data,
		Mode:    mode,
		Armored: armored,
	})
	require.EqualError(t, err, "no decryption key found for message")

	_, err = aliceService.Encrypt(context.TODO(), &EncryptRequest{
		Data:    []byte(message),
		Sender:  alice.ID().String(),
		Mode:    mode,
		Armored: armored,
	})
	require.EqualError(t, err, "no recipients specified")

	// Decrypt garbage
	_, err = aliceService.Decrypt(context.TODO(), &DecryptRequest{
		Data:    []byte("????"),
		Mode:    mode,
		Armored: armored,
	})
	require.EqualError(t, err, "invalid data")

	// Decrypt empty
	_, err = aliceService.Decrypt(context.TODO(), &DecryptRequest{
		Data:    []byte{},
		Mode:    mode,
		Armored: armored,
	})
	require.EqualError(t, err, "invalid data")
}

func TestEncryptAnonymous(t *testing.T) {
	env := newTestEnv(t)

	aliceService, aliceCloseFn := newTestService(t, env)
	defer aliceCloseFn()
	testAuthSetup(t, aliceService)
	testImportKey(t, aliceService, alice)

	bobService, bobCloseFn := newTestService(t, env)
	defer bobCloseFn()
	testAuthSetup(t, bobService)
	testImportKey(t, bobService, bob)

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

func TestEncryptDecryptStream(t *testing.T) {
	env := newTestEnv(t)
	aliceService, aliceCloseFn := newTestService(t, env)
	defer aliceCloseFn()
	testAuthSetup(t, aliceService)
	testImportKey(t, aliceService, alice)

	bobService, bobCloseFn := newTestService(t, env)
	defer bobCloseFn()
	testAuthSetup(t, bobService)
	testImportKey(t, bobService, bob)

	// Save alice's sign public key to bob's keystore, check for box key conversion.
	err := bobService.ks.SaveSignPublicKey(alice.PublicKey())
	require.NoError(t, err)
	sk, err := bobService.ks.FindEd25519PublicKey(alice.Curve25519Key().PublicKey())
	require.NoError(t, err)
	require.NotNil(t, sk)
	require.Equal(t, alice.ID(), sk.ID())

	testEncryptDecryptStream(t, aliceService, bobService, bytes.Repeat([]byte{0x31}, 5), alice.ID().String(), []string{bob.ID().String()}, SigncryptV1, true)
	testEncryptDecryptStream(t, aliceService, bobService, bytes.Repeat([]byte{0x31}, 5), alice.ID().String(), []string{bob.ID().String()}, SigncryptV1, false)
	testEncryptDecryptStream(t, aliceService, bobService, bytes.Repeat([]byte{0x31}, 5), alice.ID().String(), []string{bob.ID().String()}, EncryptV2, true)
	testEncryptDecryptStream(t, aliceService, bobService, bytes.Repeat([]byte{0x31}, 5), alice.ID().String(), []string{bob.ID().String()}, EncryptV2, false)
	testEncryptDecryptStream(t, aliceService, bobService, bytes.Repeat([]byte{0x31}, (1024*1024)+5), alice.ID().String(), []string{bob.ID().String()}, SigncryptV1, true)
	testEncryptDecryptStream(t, aliceService, bobService, bytes.Repeat([]byte{0x31}, (1024*1024)+5), alice.ID().String(), []string{bob.ID().String()}, SigncryptV1, false)
	testEncryptDecryptStream(t, aliceService, bobService, bytes.Repeat([]byte{0x31}, (1024*1024)+5), alice.ID().String(), []string{bob.ID().String()}, EncryptV2, true)
	testEncryptDecryptStream(t, aliceService, bobService, bytes.Repeat([]byte{0x31}, (1024*1024)+5), alice.ID().String(), []string{bob.ID().String()}, EncryptV2, false)
	// TODO: Test timeout if data stops streaming
}

func testEncryptDecryptStream(t *testing.T, aliceService *service, bobService *service, plaintext []byte, sender string, recipients []string, mode EncryptMode, armored bool) {
	encrypted, err := testEncryptStream(t, aliceService, plaintext, sender, recipients, mode, armored)
	require.NoError(t, err)

	out, signer, err := testDecryptStream(t, bobService, encrypted, mode, armored)
	require.NoError(t, err)
	require.Equal(t, plaintext, out)
	require.Equal(t, sender, signer.ID)
}

func testEncryptStream(t *testing.T, service *service, plaintext []byte, sender string, recipients []string, mode EncryptMode, armored bool) ([]byte, error) {
	client, clientCloseFn := newTestRPCClient(t, service)
	defer clientCloseFn()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	streamClient, streamErr := client.ProtoClient().EncryptStream(ctx)
	require.NoError(t, streamErr)

	chunkSize := 1024 * 1024
	go func() {
		done := false
		err := streamClient.Send(&EncryptStreamInput{
			Recipients: recipients,
			Sender:     sender,
			Armored:    armored,
			Mode:       mode,
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

	return buf.Bytes(), nil
}

func testDecryptStream(t *testing.T, service *service, b []byte, mode EncryptMode, armored bool) ([]byte, *Key, error) {
	sp := saltpack.NewSaltpack(service.ks)
	sp.SetArmored(armored)

	client, clientCloseFn := newTestRPCClient(t, service)
	defer clientCloseFn()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	chunkSize := 1024 * 1024

	var streamClient DecryptStreamClient
	var clientErr error
	switch mode {
	case SigncryptV1:
		if armored {
			streamClient, clientErr = client.ProtoClient().SigncryptOpenArmoredStream(ctx)
		} else {
			streamClient, clientErr = client.ProtoClient().SigncryptOpenStream(ctx)
		}
	case EncryptV2:
		if armored {
			streamClient, clientErr = client.ProtoClient().DecryptArmoredStream(ctx)
		} else {
			streamClient, clientErr = client.ProtoClient().DecryptStream(ctx)
		}
	default:
		t.Fatal("invalid mode")
	}
	if clientErr != nil {
		return nil, nil, clientErr
	}

	go func() {
		done := false
		for chunk := 0; ; chunk++ {
			s, e := (chunk * chunkSize), ((chunk + 1) * chunkSize)
			if e > len(b) {
				e = len(b)
				done = true
			}
			err := streamClient.Send(&DecryptStreamInput{
				Data: b[s:e],
			})
			require.NoError(t, err)
			if done {
				break
			}
		}
		closeErr := streamClient.CloseSend()
		require.NoError(t, closeErr)
	}()

	var bufOut bytes.Buffer
	var sender *Key
	for {
		resp, recvErr := streamClient.Recv()
		if recvErr != nil {
			return nil, nil, recvErr
		}
		if len(resp.Data) == 0 {
			break
		}
		_, writeErr := bufOut.Write(resp.Data)
		require.NoError(t, writeErr)
		sender = resp.Sender
	}

	return bufOut.Bytes(), sender, nil
}

func TestDecryptStreamInvalid(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	testAuthSetup(t, service)
	testImportKey(t, service, alice)
	testImportKey(t, service, bob)

	_, _, err := testDecryptStream(t, service, []byte("???"), SigncryptV1, true)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, "unexpected EOF", st.Message())
}
