package service

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/request"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/status"
)

func TestEncryptDecrypt(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// keys.SetLogger(NewLogger(DebugLevel))
	// saltpack.SetLogger(NewLogger(DebugLevel))
	env := newTestEnv(t)

	aliceService, aliceCloseFn := newTestService(t, env, "")
	defer aliceCloseFn()
	testAuthSetup(t, aliceService)
	testImportKey(t, aliceService, alice)

	bobService, bobCloseFn := newTestService(t, env, "")
	defer bobCloseFn()
	testAuthSetup(t, bobService)
	testImportKey(t, bobService, bob)
	testImportID(t, bobService, alice.ID())

	testEncryptDecrypt(t, aliceService, bobService, alice.ID().String(), bob.ID().String(), DefaultEncrypt, false, alice.ID())
	testEncryptDecrypt(t, aliceService, bobService, alice.ID().String(), bob.ID().String(), DefaultEncrypt, true, alice.ID())
	testEncryptDecrypt(t, aliceService, bobService, alice.ID().String(), bob.ID().String(), SaltpackEncrypt, false, alice.ID())
	testEncryptDecrypt(t, aliceService, bobService, alice.ID().String(), bob.ID().String(), SaltpackEncrypt, true, alice.ID())
	testEncryptDecrypt(t, aliceService, bobService, alice.ID().String(), bob.ID().String(), SaltpackSigncrypt, false, alice.ID())
	testEncryptDecrypt(t, aliceService, bobService, alice.ID().String(), bob.ID().String(), SaltpackSigncrypt, true, alice.ID())

	testEncryptDecryptErrors(t, aliceService, bobService, DefaultEncrypt, false)
	testEncryptDecryptErrors(t, aliceService, bobService, DefaultEncrypt, true)
	testEncryptDecryptErrors(t, aliceService, bobService, SaltpackEncrypt, false)
	testEncryptDecryptErrors(t, aliceService, bobService, SaltpackEncrypt, true)
	testEncryptDecryptErrors(t, aliceService, bobService, SaltpackSigncrypt, false)
	testEncryptDecryptErrors(t, aliceService, bobService, SaltpackSigncrypt, true)
}

func testEncryptDecrypt(t *testing.T, aliceService *service, bobService *service, sender string, recipient string, mode EncryptMode, armored bool, expectedSigner keys.ID) {
	message := "Hey bob"

	// Encrypt
	encryptResp, err := aliceService.Encrypt(context.TODO(), &EncryptRequest{
		Data:       []byte(message),
		Recipients: []string{recipient},
		Sender:     sender,
		Options: &EncryptOptions{
			Mode: mode,
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, encryptResp.Data)

	// Decrypt
	decryptResp, err := bobService.Decrypt(context.TODO(), &DecryptRequest{
		Data: encryptResp.Data,
	})
	require.NoError(t, err)
	require.Equal(t, message, string(decryptResp.Data))
	require.Equal(t, expectedSigner.String(), decryptResp.Sender.ID)

	// Decrypt (alice)
	decryptResp, err = aliceService.Decrypt(context.TODO(), &DecryptRequest{
		Data: encryptResp.Data,
	})
	require.NoError(t, err)
	require.Equal(t, message, string(decryptResp.Data))
	require.Equal(t, expectedSigner.String(), decryptResp.Sender.ID)
}

func testEncryptDecryptErrors(t *testing.T, aliceService *service, bobService *service, mode EncryptMode, armored bool) {
	message := "Hey bob"

	encryptResp, err := aliceService.Encrypt(context.TODO(), &EncryptRequest{
		Data:       []byte(message),
		Recipients: []string{bob.ID().String()},
		Sender:     alice.ID().String(),
		Options: &EncryptOptions{
			Mode:              mode,
			Armored:           armored,
			NoSenderRecipient: true,
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, encryptResp.Data)

	// Alice try to decrypt her own message
	_, err = aliceService.Decrypt(context.TODO(), &DecryptRequest{
		Data: encryptResp.Data,
	})
	require.EqualError(t, err, "no decryption key found for message")

	_, err = aliceService.Encrypt(context.TODO(), &EncryptRequest{
		Data: []byte(message),
	})
	require.EqualError(t, err, "no recipients specified")

	// Decrypt garbage
	_, err = aliceService.Decrypt(context.TODO(), &DecryptRequest{
		Data: []byte("????"),
	})
	require.EqualError(t, err, "invalid data")

	// Decrypt empty
	_, err = aliceService.Decrypt(context.TODO(), &DecryptRequest{
		Data: []byte{},
	})
	require.EqualError(t, err, "invalid data")
}

func TestEncryptAnonymous(t *testing.T) {
	env := newTestEnv(t)

	aliceService, aliceCloseFn := newTestService(t, env, "")
	defer aliceCloseFn()
	testAuthSetup(t, aliceService)
	testImportKey(t, aliceService, alice)

	bobService, bobCloseFn := newTestService(t, env, "")
	defer bobCloseFn()
	testAuthSetup(t, bobService)
	testImportKey(t, bobService, bob)

	message := "Hey bob"

	// Encrypt
	encryptResp, err := aliceService.Encrypt(context.TODO(), &EncryptRequest{
		Data:       []byte(message),
		Recipients: []string{bob.ID().String()},
		Sender:     "",
		Options: &EncryptOptions{
			Mode: SaltpackEncrypt,
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, encryptResp.Data)

	// Decrypt
	decryptResp, err := bobService.Decrypt(context.TODO(), &DecryptRequest{
		Data: encryptResp.Data,
	})
	require.NoError(t, err)
	require.Equal(t, message, string(decryptResp.Data))
	require.Nil(t, decryptResp.Sender)

	// Encrypt (signcrypt)
	encryptResp, err = aliceService.Encrypt(context.TODO(), &EncryptRequest{
		Data:       []byte(message),
		Recipients: []string{bob.ID().String()},
		Sender:     "",
		Options: &EncryptOptions{
			Mode: SaltpackSigncrypt,
		},
	})
	// Decrypt
	decryptResp, err = bobService.Decrypt(context.TODO(), &DecryptRequest{
		Data: encryptResp.Data,
	})
	require.NoError(t, err)
	require.Equal(t, message, string(decryptResp.Data))
	require.Nil(t, decryptResp.Sender)

	// Encrypt (no sign)
	encryptResp, err = aliceService.Encrypt(context.TODO(), &EncryptRequest{
		Data:       []byte(message),
		Recipients: []string{bob.ID().String()},
		Sender:     alice.ID().String(),
		Options: &EncryptOptions{
			NoSign: true,
		},
	})

	// Decrypt
	decryptResp, err = bobService.Decrypt(context.TODO(), &DecryptRequest{
		Data: encryptResp.Data,
	})
	require.NoError(t, err)
	require.Equal(t, message, string(decryptResp.Data))
	require.Nil(t, decryptResp.Sender)
}

func TestEncryptDecryptStream(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))

	env := newTestEnv(t)
	aliceService, aliceCloseFn := newTestService(t, env, "")
	defer aliceCloseFn()
	testAuthSetup(t, aliceService)
	testImportKey(t, aliceService, alice)

	bobService, bobCloseFn := newTestService(t, env, "")
	defer bobCloseFn()
	testAuthSetup(t, bobService)
	testImportKey(t, bobService, bob)

	testImportID(t, bobService, alice.ID())

	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, 5), alice.ID().String(), []string{bob.ID().String()}, DefaultEncrypt, false, alice.ID())
	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, 5), alice.ID().String(), []string{bob.ID().String()}, SaltpackEncrypt, false, alice.ID())
	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, 5), alice.ID().String(), []string{bob.ID().String()}, SaltpackEncrypt, true, alice.ID())
	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, 5), alice.ID().String(), []string{bob.ID().String()}, SaltpackSigncrypt, false, alice.ID())
	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, 5), alice.ID().String(), []string{bob.ID().String()}, SaltpackSigncrypt, true, alice.ID())
	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, (1024*1024)+5), alice.ID().String(), []string{bob.ID().String()}, DefaultEncrypt, false, alice.ID())
	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, (1024*1024)+5), alice.ID().String(), []string{bob.ID().String()}, SaltpackEncrypt, false, alice.ID())
	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, (1024*1024)+5), alice.ID().String(), []string{bob.ID().String()}, SaltpackEncrypt, true, alice.ID())
	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, (1024*1024)+5), alice.ID().String(), []string{bob.ID().String()}, SaltpackSigncrypt, false, alice.ID())
	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, (1024*1024)+5), alice.ID().String(), []string{bob.ID().String()}, SaltpackSigncrypt, true, alice.ID())
	// TODO: Test timeout if data stops streaming
}

func testEncryptDecryptStream(t *testing.T, env *testEnv,
	aliceService *service, bobService *service,
	plaintext []byte, sender string, recipients []string,
	mode EncryptMode, armored bool, expectedSender keys.ID) {
	encrypted, err := testEncryptStream(t, env, aliceService, plaintext, sender, recipients, mode, armored)
	require.NoError(t, err)

	out, outSigner, outMode, err := testDecryptStream(t, env, bobService, encrypted)
	require.NoError(t, err)
	require.Equal(t, plaintext, out)
	if mode == DefaultEncrypt {
		require.Equal(t, SaltpackSigncrypt, outMode)
	} else {
		require.Equal(t, mode, outMode)
	}
	require.NotNil(t, outSigner)
	require.Equal(t, expectedSender.String(), outSigner.ID)
}

func testEncryptStream(t *testing.T, env *testEnv, service *service, plaintext []byte, sender string, recipients []string, mode EncryptMode, armored bool) ([]byte, error) {
	// TODO: Assert out
	client, clientCloseFn := newTestRPCClient(t, service, env, "", nil)
	defer clientCloseFn()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	streamClient, streamErr := client.KeysClient().EncryptStream(ctx)
	require.NoError(t, streamErr)

	chunkSize := 1024 * 1024
	go func() {
		done := false
		err := streamClient.Send(&EncryptInput{
			Recipients: recipients,
			Sender:     sender,
			Options: &EncryptOptions{
				Mode: mode,
			},
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

func testDecryptStream(t *testing.T, env *testEnv, service *service, b []byte) ([]byte, *Key, EncryptMode, error) {
	// TODO: Assert out
	client, clientCloseFn := newTestRPCClient(t, service, env, "", nil)
	defer clientCloseFn()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	chunkSize := 1024 * 1024

	streamClient, err := client.KeysClient().DecryptStream(ctx)
	if err != nil {
		return nil, nil, DefaultEncrypt, err
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
	var mode EncryptMode
	for {
		resp, recvErr := streamClient.Recv()
		if recvErr != nil {
			return nil, nil, DefaultEncrypt, recvErr
		}
		_, writeErr := bufOut.Write(resp.Data)
		require.NoError(t, writeErr)
		if resp.Sender != nil {
			sender = resp.Sender
		}
		mode = resp.Mode
		if len(resp.Data) == 0 {
			break
		}
	}

	return bufOut.Bytes(), sender, mode, nil
}

func TestDecryptStreamInvalid(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()
	testAuthSetup(t, service)
	testImportKey(t, service, alice)
	testImportKey(t, service, bob)

	_, _, _, err := testDecryptStream(t, env, service, []byte("???"))
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, "invalid data", st.Message())
}

func TestEncryptDecryptByUser(t *testing.T) {
	env := newTestEnv(t)

	aliceService, aliceCloseFn := newTestService(t, env, "")
	defer aliceCloseFn()
	testAuthSetup(t, aliceService)
	testImportKey(t, aliceService, alice)
	testUserSetupGithub(t, env, aliceService, alice, "alice")

	bobService, bobCloseFn := newTestService(t, env, "")
	defer bobCloseFn()
	testAuthSetup(t, bobService)
	testImportKey(t, bobService, bob)
	testUserSetupGithub(t, env, bobService, bob, "bob")

	testPull(t, aliceService, bob.ID())
	testPull(t, bobService, alice.ID())

	testEncryptDecrypt(t, aliceService, bobService, "alice@github", "bob@github", SaltpackEncrypt, false, alice.ID())
	testEncryptDecrypt(t, aliceService, bobService, "alice@github", "bob@github", SaltpackSigncrypt, false, alice.ID())

	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, (1024*1024)+5), "alice@github", []string{"bob@github"}, SaltpackEncrypt, false, alice.ID())
	testEncryptDecryptStream(t, env, aliceService, bobService, bytes.Repeat([]byte{0x31}, (1024*1024)+5), "alice@github", []string{"bob@github"}, SaltpackSigncrypt, false, alice.ID())
}

func TestEncryptDecryptFile(t *testing.T) {
	env := newTestEnv(t)

	aliceService, aliceCloseFn := newTestService(t, env, "")
	defer aliceCloseFn()
	testAuthSetup(t, aliceService)
	testImportKey(t, aliceService, alice)

	bobService, bobCloseFn := newTestService(t, env, "")
	defer bobCloseFn()
	testAuthSetup(t, bobService)
	testImportKey(t, bobService, bob)

	testImportID(t, bobService, alice.ID())

	b := []byte("test message")
	inPath := keys.RandTempPath()
	encPath := inPath + ".enc"
	decPath := inPath + ".dec"

	defer os.Remove(inPath)
	defer os.Remove(encPath)
	defer os.Remove(decPath)

	writeErr := ioutil.WriteFile(inPath, b, 0644)
	require.NoError(t, writeErr)

	aliceClient, aliceClientCloseFn := newTestRPCClient(t, aliceService, env, "", nil)
	defer aliceClientCloseFn()

	options := &EncryptOptions{}
	err := encryptFile(aliceClient, inPath, encPath, []string{bob.ID().String()}, alice.ID().String(), options)
	require.NoError(t, err)

	// encrypted, err := ioutil.ReadFile(outPath)
	// require.NoError(t, err)
	// t.Logf("encrypted: %s", string(encrypted))

	bobClient, bobClientCloseFn := newTestRPCClient(t, bobService, env, "", nil)
	defer bobClientCloseFn()

	dec, err := decryptFile(bobClient, encPath, decPath)
	require.NoError(t, err)
	require.NotNil(t, dec.Sender)
	require.Equal(t, alice.ID().String(), dec.Sender.ID)
	require.Equal(t, decPath, dec.Out)

	bout, err := ioutil.ReadFile(decPath)
	require.NoError(t, err)
	require.Equal(t, b, bout)
	os.Remove(decPath)

	// Decrypt .dat file
	datPath := inPath + ".dat"
	err = os.Rename(encPath, datPath)
	require.NoError(t, err)

	dec, err = decryptFile(bobClient, datPath, "")
	require.NoError(t, err)
	require.NotNil(t, dec.Sender)
	require.Equal(t, alice.ID().String(), dec.Sender.ID)

	ext := path.Ext(datPath)
	noExt := datPath[0 : len(datPath)-len(ext)]
	require.Equal(t, noExt+"-2.dat", dec.Out)
}

func TestEncryptUnverified(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// keys.SetLogger(NewLogger(DebugLevel))
	env := newTestEnv(t)

	aliceService, aliceCloseFn := newTestService(t, env, "")
	defer aliceCloseFn()
	testAuthSetup(t, aliceService)
	testImportKey(t, aliceService, alice)
	testUserSetupGithub(t, env, aliceService, alice, "alice")

	bobService, bobCloseFn := newTestService(t, env, "")
	defer bobCloseFn()
	testAuthSetup(t, bobService)
	testImportKey(t, bobService, bob)
	testUserSetupGithub(t, env, bobService, bob, "bob")

	// Encrypt (not found)
	_, err := aliceService.Encrypt(context.TODO(), &EncryptRequest{
		Data:       []byte("hi"),
		Recipients: []string{"bob@github"},
		Sender:     "alice@github",
	})
	require.EqualError(t, err, "not found bob@github")

	testPull(t, aliceService, bob.ID())

	env.clock.Add(time.Hour * 24)

	// Set 500 error for bob@github
	env.req.SetError("https://gist.github.com/bob/1", request.ErrHTTP{StatusCode: 500})

	// Encrypt (bob, error)
	_, err = aliceService.Encrypt(context.TODO(), &EncryptRequest{
		Data:       []byte("hi"),
		Recipients: []string{"bob@github"},
		Sender:     "alice@github",
	})
	require.EqualError(t, err, "user bob@github has failed status connection-fail")
}
