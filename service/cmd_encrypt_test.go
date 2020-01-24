package service

// func TestEncryptCommand(t *testing.T) {
// 	SetLog(newLog(DebugLevel))
// 	service := testService(t)
// 	defer service.Close()
// 	client, closeFn := newTestClient(t, service)
// 	defer closeFn()

// 	err := service.auth.Setup("testtoken123")
// 	require.NoError(t, err)
// 	err := os.Setenv("KEYS_AUTH", "testtoken123")
// 	require.NoError(t, err)

// 	genResp, err := service.KeyGenerate(context.TODO(), &KeyGenerateRequest{})
// 	require.NoError(t, err)
// 	kid := genResp.UserPublicKey.ID

// 	inPath := newTempPath(t, "txt")
// 	outPath := inPath + ".enc"

// 	var clientErr error
// 	errorFn := func(err error) {
// 		clientErr = err
// 	}

// 	args := os.Args[0:1]
// 	args = append(args, "encrypt", "-recipients", kid, "-in", inPath, "-out", outPath)
// 	runClient(Build{}, args, client, errorFn)
// 	require.EqualError(t, clientErr, fmt.Sprintf("open %s: no such file or directory", inPath))
// 	clientErr = nil
// 	writeErr := ioutil.WriteFile(inPath, []byte("test message"), 0644)
// 	require.NoError(t, writeErr)

// 	runClient(Build{}, args, client, errorFn)
// 	require.NoError(t, clientErr)
// }
