package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()

	testAuthSetup(t, service)

	resp, err := service.ConfigGet(context.TODO(), &ConfigGetRequest{Name: "encrypt"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Nil(t, resp.Config)

	config := &Config{
		Encrypt: &Config_Encrypt{
			Recipients:        []string{"gabriel@github"},
			Sender:            "gabriel@echo",
			NoSenderRecipient: true,
			NoSign:            true,
		},
	}
	_, err = service.ConfigSet(context.TODO(), &ConfigSetRequest{
		Name:   "encrypt",
		Config: config,
	})
	require.NoError(t, err)

	resp, err = service.ConfigGet(context.TODO(), &ConfigGetRequest{Name: "encrypt"})
	require.NoError(t, err)
	require.NotNil(t, resp.Config)
	require.NotNil(t, resp.Config.Encrypt)
	encrypt := config.Encrypt
	out := resp.Config.Encrypt
	require.Equal(t, encrypt.Recipients, out.Recipients)
	require.Equal(t, encrypt.Sender, out.Sender)
	require.Equal(t, encrypt.NoSenderRecipient, out.NoSenderRecipient)
	require.Equal(t, encrypt.NoSign, out.NoSign)
}
