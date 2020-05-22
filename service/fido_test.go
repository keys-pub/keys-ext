package service

// func goBin(t *testing.T) string {
// 	usr, err := user.Current()
// 	require.NoError(t, err)
// 	return filepath.Join(usr.HomeDir, "go", "bin")
// }

// func TestHMACSecretAuth(t *testing.T) {
// 	cfg, closeFn := testConfig(t, "KeysTest", "", "mem")
// 	defer closeFn()
// 	st, err := newKeyringStore(cfg)
// 	require.NoError(t, err)
// 	auth, err := newAuth(cfg, st)
// 	require.NoError(t, err)

// 	fido2Plugin, err := fido2.OpenPlugin(filepath.Join(goBin(t), "fido2.so"))
// 	require.NoError(t, err)
// 	auth.authenticators = fido2Plugin

// 	setup, err := auth.unlock(context.TODO(), "12345", FIDO2HMACSecretAuth, "test", true)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, setup.token)

// 	unlock, err := auth.unlock(context.TODO(), "12345", FIDO2HMACSecretAuth, "test", false)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, unlock.token)
// 	require.Equal(t, setup.auth.Key(), unlock.auth.Key())
// }
