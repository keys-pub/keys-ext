package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrefs(t *testing.T) {
	ctx := context.TODO()
	env := newTestEnv(t)

	service, serviceCloseFn := newTestService(t, env, "")
	defer serviceCloseFn()
	testAuthSetup(t, service)

	resp, err := service.Preferences(ctx, &PreferencesRequest{})
	require.NoError(t, err)
	require.Equal(t, []*Pref{}, resp.Prefs)

	_, err = service.PreferenceSet(ctx, &PreferenceSetRequest{
		Pref: &Pref{Key: "key1", Value: "value1"},
	})
	require.NoError(t, err)

	resp, err = service.Preferences(ctx, &PreferencesRequest{})
	require.NoError(t, err)
	require.Equal(t, []*Pref{&Pref{Key: "key1", Value: "value1"}}, resp.Prefs)
}
