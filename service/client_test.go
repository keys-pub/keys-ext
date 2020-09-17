package service

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type listener struct {
	lis *bufconn.Listener
}

func (l listener) dial(context.Context, string) (net.Conn, error) {
	return l.lis.Dial()
}

func newTestRPCClient(t *testing.T, srvc *service, tenv *testEnv, appName string, out io.Writer) (*Client, func()) {
	if appName == "" {
		appName = "KeysTest-" + randName()
	}
	listener := listener{lis: bufconn.Listen(1024 * 1024)}

	connect := func(env *Env, authToken string) (*grpc.ClientConn, error) {
		logger.Debugf("Test connect %s", authToken)
		var opts []grpc.DialOption
		opts = append(opts, grpc.WithContextDialer(listener.dial), grpc.WithInsecure())
		opts = append(opts, grpc.WithPerRPCCredentials(newTestClientAuth(authToken)))
		return grpc.DialContext(context.TODO(), "bufnet", opts...)
	}

	server := grpc.NewServer()
	RegisterKeysServer(server, srvc)
	go func() {
		serveErr := server.Serve(listener.lis)
		require.NoError(t, serveErr)
	}()

	client := NewClient()
	client.connectFn = connect
	if out != nil {
		client.out = out
	}
	env, closeFn := newEnv(t, appName, "")
	err := client.Connect(env, "")
	require.NoError(t, err)

	closeClientFn := func() {
		// TODO: Remove sleep
		time.Sleep(time.Millisecond * 100)
		client.Close()
		server.Stop()
		closeFn()
	}

	return client, closeClientFn
}
