package service

import (
	"context"
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

func newTestRPCClient(t *testing.T, srvc *service) (*Client, func()) {
	listener := listener{lis: bufconn.Listen(1024 * 1024)}

	connect := func(cfg *Config, authToken string) (*grpc.ClientConn, error) {
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
	cfg, cfgClose := testConfig(t, "")
	err := client.Connect(cfg, "")
	require.NoError(t, err)

	closeFn := func() {
		// TODO: Remove sleep
		time.Sleep(time.Millisecond * 100)
		client.Close()
		server.Stop()
		cfgClose()
	}

	return client, closeFn
}
