package service

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/keys-pub/keys/saltpack"
	"github.com/stretchr/testify/require"
)

func TestSignVerify(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	testAuthSetup(t, service, alice)

	message := "I'm alice"
	signResp, err := service.Sign(context.TODO(), &SignRequest{Data: []byte(message), KID: alice.ID().String()})
	require.NoError(t, err)
	require.NotEmpty(t, signResp.Data)
	require.Equal(t, alice.ID().String(), signResp.KID)

	verifyResp, err := service.Verify(context.TODO(), &VerifyRequest{Data: signResp.Data})
	require.NoError(t, err)
	require.Equal(t, message, string(verifyResp.Data))
	require.Equal(t, alice.ID().String(), verifyResp.KID)

	signResp, err = service.Sign(context.TODO(), &SignRequest{Data: []byte(message)})
	require.NoError(t, err)
	require.NotEmpty(t, signResp.Data)
	require.Equal(t, alice.ID().String(), signResp.KID)
}

func TestSignStream(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	testAuthSetup(t, service, alice)

	testSignStream(t, service, bytes.Repeat([]byte{0x31}, 5), alice.ID().String())
	testSignStream(t, service, bytes.Repeat([]byte{0x31}, 5), "")
	testSignStream(t, service, bytes.Repeat([]byte{0x31}, (1024*1024)+5), alice.ID().String())
	// TODO: Test timeout if data stops streaming
}

func testSignStream(t *testing.T, service *service, plaintext []byte, sender string) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	cl, clientCloseFn := newTestRPCClient(t, service)
	defer clientCloseFn()

	streamClient, streamErr := cl.ProtoClient().SignStream(ctx)
	require.NoError(t, streamErr)

	chunkSize := 1024 * 1024
	go func() {
		done := false
		err := streamClient.Send(&SignStreamInput{
			KID:     sender,
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
			err := streamClient.Send(&SignStreamInput{
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
	sp.SetArmored(true)
	out, sout, err := sp.Verify(data)
	require.NoError(t, err)
	if sender != "" {
		require.Equal(t, sout.String(), sender)
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
			err := outClient.Send(&VerifyStreamInput{
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
