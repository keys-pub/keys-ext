package client

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestPubSub(t *testing.T) {
	t.Skip()
	// SetLogger(NewLogger(DebugLevel))
	// api.SetLogger(NewLogger(DebugLevel))
	// server.SetContextLogger(NewContextLogger(DebugLevel))

	env := testEnv(t, logger)
	defer env.closeFn()

	ksa := keys.NewMemKeystore()
	aliceClient := testClient(t, env, ksa)
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	err := ksa.SaveEdX25519Key(alice)
	require.NoError(t, err)

	ksb := keys.NewMemKeystore()
	bobClient := testClient(t, env, ksb)
	bob := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))
	err = ksb.SaveEdX25519Key(bob)
	require.NoError(t, err)

	// Pub
	err = aliceClient.Publish(context.TODO(), alice.ID(), bob.ID(), []byte("hi"), MessagePub)
	require.NoError(t, err)

	// Sub
	wg := &sync.WaitGroup{}
	wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	msgs := []string{}
	receiveFn := func(msg *PubSubMessage) {
		msgs = append(msgs, string(msg.Data))
		require.Equal(t, alice.ID(), msg.Sender)
		require.Equal(t, MessagePub, msg.Type)
		t.Logf("msg: %v", msg)
		if len(msgs) >= 2 {
			cancel()
		}
	}
	go func() {
		err := bobClient.Subscribe(ctx, bob.ID(), receiveFn)
		require.NoError(t, err)
		wg.Done()
	}()

	// Pub
	err = aliceClient.Publish(context.TODO(), alice.ID(), bob.ID(), []byte("what time is the meeting?"), MessagePub)
	require.NoError(t, err)

	wg.Wait()

	require.Equal(t, []string{"hi", "what time is the meeting?"}, msgs)

}

func TestWebsocket(t *testing.T) {
	t.Skip()
	// url := fmt.Sprintf("wss://keys.pub/wsecho")
	// url := fmt.Sprintf("ws://localhost:8080/wsecho")
	url := fmt.Sprintf("ws://echo.websocket.org")

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	// t.Logf("resp: %+v", resp)
	// body, _ := ioutil.ReadAll(resp.Body)
	// t.Logf("body: %s", string(body))
	require.NoError(t, err)
	defer conn.Close()

	err = conn.WriteMessage(websocket.TextMessage, []byte("ping"))
	require.NoError(t, err)

	_, b, err := conn.ReadMessage()
	require.NoError(t, err)
	require.Equal(t, b, []byte("ping"))

	logger.Infof("msg: %+v\n", b)
}
