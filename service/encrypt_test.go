package service

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/status"
)

func TestEncryptDecrypt(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// keys.SetLogger(NewLogger(DebugLevel))
	// saltpack.SetLogger(NewLogger(DebugLevel))
	env := newTestEnv(t)

	aliceService, aliceCloseFn := newTestService(t, env)
	defer aliceCloseFn()
	testAuthSetup(t, aliceService)
	testImportKey(t, aliceService, alice)

	bobService, bobCloseFn := newTestService(t, env)
	defer bobCloseFn()
	testAuthSetup(t, bobService)
	testImportKey(t, bobService, bob)
	testImportID(t, bobService, alice.ID())

	testEncryptDecrypt(t, aliceService, bobService, alice.ID().String(), bob.ID().String(), DefaultEncryptMode, true, alice.ID())
	testEncryptDecrypt(t, aliceService, bobService, alice.ID().String(), bob.ID().String(), DefaultEncryptMode, false, alice.ID())
	testEncryptDecrypt(t, aliceService, bobService, alice.ID().String(), bob.ID().String(), EncryptV2, true, alice.ID())
	testEncryptDecrypt(t, aliceService, bobService, alice.ID().String(), bob.ID().String(), EncryptV2, false, alice.ID())
	testEncryptDecrypt(t, aliceService, bobService, alice.ID().String(), bob.ID().String(), SigncryptV1, true, alice.ID())
	testEncryptDecrypt(t, aliceService, bobService, alice.ID().String(), bob.ID().String(), SigncryptV1, false, alice.ID())

	testEncryptDecryptErrors(t, aliceService, bobService, DefaultEncryptMode, true)
	testEncryptDecryptErrors(t, aliceService, bobService, DefaultEncryptMode, false)
	testEncryptDecryptErrors(t, aliceService, bobService, EncryptV2, true)
	testEncryptDecryptErrors(t, aliceService, bobService, EncryptV2, false)
	testEncryptDecryptErrors(t, aliceService, bobService, SigncryptV1, true)
	testEncryptDecryptErrors(t, aliceService, bobService, SigncryptV1, false)
}

func testEncryptDecrypt(t *testing.T, aliceService *service, bobService *service, sender string, recipient string, mode EncryptMode, armored bool, expectedSigner keys.ID) {
	message := "Hey bob"

	// Encrypt
	encryptResp, err := aliceService.Encrypt(context.TODO(), &EncryptRequest{
		Data:       []byte(message),
		Sender:     sender,
		Recipients: []string{recipient, sender},
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
	require.Equal(t, expectedSigner.String(), decryptResp.Sender.ID)

	// Decrypt (alice)
	decryptResp, err = aliceService.Decrypt(context.TODO(), &DecryptRequest{
		Data:    encryptResp.Data,
		Mode:    mode,
		Armored: armored,
	})
	require.NoError(t, err)
	require.Equal(t, message, string(decryptResp.Data))
	require.Equal(t, expectedSigner.String(), decryptResp.Sender.ID)
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
	if armored {
		require.EqualError(t, err, "unexpected EOF")
	} else {
		require.EqualError(t, err, "invalid data")
	}

	// Decrypt empty
	_, err = aliceService.Decrypt(context.TODO(), &DecryptRequest{
		Data:    []byte{},
		Mode:    mode,
		Armored: armored,
	})
	if armored {
		require.EqualError(t, err, "unexpected EOF")
	} else {
		require.EqualError(t, err, "invalid data")
	}
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
	// SetLogger(NewLogger(DebugLevel))

	env := newTestEnv(t)
	aliceService, aliceCloseFn := newTestService(t, env)
	defer aliceCloseFn()
	testAuthSetup(t, aliceService)
	testImportKey(t, aliceService, alice)

	bobService, bobCloseFn := newTestService(t, env)
	defer bobCloseFn()
	testAuthSetup(t, bobService)
	testImportKey(t, bobService, bob)

	testImportID(t, bobService, alice.ID())

	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, 5), alice.ID().String(), []string{bob.ID().String()}, DefaultEncryptMode, true, alice.ID())
	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, 5), alice.ID().String(), []string{bob.ID().String()}, DefaultEncryptMode, false, alice.ID())
	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, 5), alice.ID().String(), []string{bob.ID().String()}, EncryptV2, true, alice.ID())
	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, 5), alice.ID().String(), []string{bob.ID().String()}, EncryptV2, false, alice.ID())
	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, 5), alice.ID().String(), []string{bob.ID().String()}, SigncryptV1, true, alice.ID())
	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, 5), alice.ID().String(), []string{bob.ID().String()}, SigncryptV1, false, alice.ID())
	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, (1024*1024)+5), alice.ID().String(), []string{bob.ID().String()}, DefaultEncryptMode, true, alice.ID())
	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, (1024*1024)+5), alice.ID().String(), []string{bob.ID().String()}, DefaultEncryptMode, false, alice.ID())
	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, (1024*1024)+5), alice.ID().String(), []string{bob.ID().String()}, EncryptV2, true, alice.ID())
	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, (1024*1024)+5), alice.ID().String(), []string{bob.ID().String()}, EncryptV2, false, alice.ID())
	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, (1024*1024)+5), alice.ID().String(), []string{bob.ID().String()}, SigncryptV1, true, alice.ID())
	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, (1024*1024)+5), alice.ID().String(), []string{bob.ID().String()}, SigncryptV1, false, alice.ID())
	// TODO: Test timeout if data stops streaming
}

func testEncryptDecryptStream(t *testing.T, env *testEnv, aliceService *service, bobService *service, plaintext []byte, sender string, recipients []string, mode EncryptMode, armored bool, expectedSender keys.ID) {
	encrypted, err := testEncryptStream(t, env, aliceService, plaintext, sender, recipients, mode, armored)
	require.NoError(t, err)

	if mode == DefaultEncryptMode {
		mode = EncryptV2
	}

	out, outSigner, err := testDecryptStream(t, env, bobService, encrypted, mode, armored)
	require.NoError(t, err)
	require.Equal(t, plaintext, out)
	require.NotNil(t, outSigner)
	require.Equal(t, expectedSender.String(), outSigner.ID)
}

func testEncryptStream(t *testing.T, env *testEnv, service *service, plaintext []byte, sender string, recipients []string, mode EncryptMode, armored bool) ([]byte, error) {
	client, clientCloseFn := newTestRPCClient(t, service, env)
	defer clientCloseFn()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	streamClient, streamErr := client.ProtoClient().EncryptStream(ctx)
	require.NoError(t, streamErr)

	chunkSize := 1024 * 1024
	go func() {
		done := false
		err := streamClient.Send(&EncryptInput{
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
			err := streamClient.Send(&EncryptInput{
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

func testDecryptStream(t *testing.T, env *testEnv, service *service, b []byte, mode EncryptMode, armored bool) ([]byte, *Key, error) {
	client, clientCloseFn := newTestRPCClient(t, service, env)
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
			err := streamClient.Send(&DecryptInput{
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
		_, writeErr := bufOut.Write(resp.Data)
		require.NoError(t, writeErr)
		if resp.Sender != nil {
			sender = resp.Sender
		}
		if len(resp.Data) == 0 {
			break
		}
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

	_, _, err := testDecryptStream(t, env, service, []byte("???"), SigncryptV1, true)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, "unexpected EOF", st.Message())
}

func TestEncryptDecryptByUser(t *testing.T) {
	env := newTestEnv(t)

	aliceService, aliceCloseFn := newTestService(t, env)
	defer aliceCloseFn()
	testAuthSetup(t, aliceService)
	testImportKey(t, aliceService, alice)
	testUserSetupGithub(t, env, aliceService, alice, "alice")

	bobService, bobCloseFn := newTestService(t, env)
	defer bobCloseFn()
	testAuthSetup(t, bobService)
	testImportKey(t, bobService, bob)
	testUserSetupGithub(t, env, bobService, bob, "bob")

	testPull(t, aliceService, bob.ID())
	testPull(t, bobService, alice.ID())

	testEncryptDecrypt(t, aliceService, bobService, "alice@github", "bob@github", EncryptV2, true, alice.ID())
	testEncryptDecrypt(t, aliceService, bobService, "alice@github", "bob@github", SigncryptV1, true, alice.ID())

	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, (1024*1024)+5), "alice@github", []string{"bob@github"}, EncryptV2, true, alice.ID())
	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, (1024*1024)+5), "alice@github", []string{"bob@github"}, SigncryptV1, true, alice.ID())
}

func TestEncryptDecryptFile(t *testing.T) {
	env := newTestEnv(t)

	aliceService, aliceCloseFn := newTestService(t, env)
	defer aliceCloseFn()
	testAuthSetup(t, aliceService)
	testImportKey(t, aliceService, alice)

	bobService, bobCloseFn := newTestService(t, env)
	defer bobCloseFn()
	testAuthSetup(t, bobService)
	testImportKey(t, bobService, bob)

	testImportID(t, bobService, alice.ID())

	b := []byte("test message")
	inPath := keys.RandTempPath("")
	outPath := inPath + ".enc"
	decryptedPath := inPath + ".dec"

	defer os.Remove(inPath)
	defer os.Remove(outPath)
	defer os.Remove(decryptedPath)

	writeErr := ioutil.WriteFile(inPath, b, 0644)
	require.NoError(t, writeErr)

	aliceClient, aliceClientCloseFn := newTestRPCClient(t, aliceService, env)
	defer aliceClientCloseFn()

	err := encryptFile(aliceClient, []string{bob.ID().String()}, alice.ID().String(), true, EncryptV2, inPath, outPath)
	require.NoError(t, err)

	// encrypted, err := ioutil.ReadFile(outPath)
	// require.NoError(t, err)
	// t.Logf("encrypted: %s", string(encrypted))

	bobClient, bobClientCloseFn := newTestRPCClient(t, bobService, env)
	defer bobClientCloseFn()

	dec, err := decryptFile(bobClient, true, EncryptV2, outPath, decryptedPath)
	require.NoError(t, err)
	require.NotNil(t, dec.Sender)
	require.Equal(t, alice.ID().String(), dec.Sender.ID)

	bout, err := ioutil.ReadFile(decryptedPath)
	require.NoError(t, err)
	require.Equal(t, b, bout)
	os.Remove(decryptedPath)

	dec, err = decryptFile(bobClient, true, EncryptV2, outPath, "")
	require.NoError(t, err)
	require.Equal(t, inPath+"-1", dec.Out)
	os.Remove(dec.Out)
}

func TestEncryptUnverified(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// keys.SetLogger(NewLogger(DebugLevel))
	env := newTestEnv(t)

	aliceService, aliceCloseFn := newTestService(t, env)
	defer aliceCloseFn()
	testAuthSetup(t, aliceService)
	testImportKey(t, aliceService, alice)
	testUserSetupGithub(t, env, aliceService, alice, "alice")

	bobService, bobCloseFn := newTestService(t, env)
	defer bobCloseFn()
	testAuthSetup(t, bobService)
	testImportKey(t, bobService, bob)
	testUserSetupGithub(t, env, bobService, bob, "bob")

	// Encrypt (not found)
	_, err := aliceService.Encrypt(context.TODO(), &EncryptRequest{
		Data:       []byte("hi"),
		Sender:     "alice@github",
		Recipients: []string{"bob@github"},
		Mode:       EncryptV2,
	})
	require.EqualError(t, err, "not found bob@github")

	testPull(t, aliceService, bob.ID())

	env.clock.Add(time.Hour * 24)

	// Set 500 error for bob@github
	env.req.SetError("https://gist.github.com/bob/1", util.ErrHTTP{StatusCode: 500})

	// Encrypt (bob, error)
	_, err = aliceService.Encrypt(context.TODO(), &EncryptRequest{
		Data:       []byte("hi"),
		Sender:     "alice@github",
		Recipients: []string{"bob@github"},
		Mode:       EncryptV2,
	})
	require.EqualError(t, err, "user bob@github has failed status connection-fail")
}
