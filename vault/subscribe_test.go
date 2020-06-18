package vault_test

// func TestSubscribe(t *testing.T) {
// 	var err error

// 	vlt := vault.New(vault.NewMem())
// 	ch := vlt.Subscribe("test")

// 	key := keys.Rand32()
// 	id := encoding.MustEncode(bytes.Repeat([]byte{0x01}, 32), encoding.Base62)
// 	provision := &vault.Provision{ID: id}
// 	err = vlt.Setup(key, provision)
// 	require.NoError(t, err)

// 	_, err = vlt.Unlock(key)
// 	require.NoError(t, err)

// 	err = vlt.Lock()
// 	require.NoError(t, err)

// 	event := <-ch
// 	require.IsType(t, event, vault.UnlockEvent{})
// 	unlock := event.(vault.UnlockEvent)
// 	require.Equal(t, provision.ID, unlock.Provision.ID)

// 	event = <-ch
// 	require.IsType(t, event, vault.LockEvent{})

// 	_, err = vlt.Unlock(key)
// 	require.NoError(t, err)

// 	item := vault.NewItem("test", []byte("testdata"), "", time.Now())
// 	err = vlt.Set(item)
// 	require.NoError(t, err)

// 	event = <-ch
// 	require.IsType(t, event, vault.UnlockEvent{})
// 	event = <-ch
// 	require.IsType(t, event, vault.SetEvent{})
// 	create := event.(vault.SetEvent)
// 	require.Equal(t, item.ID, create.ID)

// 	vlt.Unsubscribe("test")
// }
