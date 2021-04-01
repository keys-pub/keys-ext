package server_test

// func TestBatch(t *testing.T) {
// 	env := newEnv(t)
// 	// env.logLevel = server.DebugLevel
// 	tk := testKeysSeeded()

// 	channel, channel2 := tk.channel, tk.channel2

// 	srv := newTestServerEnv(t, env)
// 	clock := env.clock

// 	reqs := api.BatchRequests{
// 		Requests: []*api.BatchRequest{},
// 	}

// 	// PUT /channel/:cid (x2)
// 	req, err := http.NewAuthRequest("PUT", dstore.Path("channel", channel.ID()), nil, "", clock.Now(), channel)
// 	require.NoError(t, err)
// 	code, _, _ := srv.Serve(req)
// 	require.Equal(t, http.StatusOK, code)
// 	req, err = http.NewAuthRequest("PUT", dstore.Path("channel", channel2.ID()), nil, "", clock.Now(), channel2)
// 	require.NoError(t, err)
// 	code, _, _ = srv.Serve(req)
// 	require.Equal(t, http.StatusOK, code)

// 	// PUT /batch
// 	req1, err := api.NewBatchRequest("1", "GET", dstore.Path("channel", channel.ID()), "", clock.Now(), channel)
// 	require.NoError(t, err)
// 	reqs.Requests = append(reqs.Requests, req1)
// 	req2, err := api.NewBatchRequest("2", "GET", dstore.Path("channel", channel2.ID()), "", clock.Now(), channel2)
// 	require.NoError(t, err)
// 	reqs.Requests = append(reqs.Requests, req2)

// 	b, err := json.Marshal(reqs)
// 	require.NoError(t, err)
// 	req, err = http.NewRequest("POST", "/batch", bytes.NewReader(b))
// 	require.NoError(t, err)
// 	code, _, body := srv.Serve(req)
// 	var resp api.BatchResponses
// 	err = json.Unmarshal([]byte(body), &resp)
// 	require.NoError(t, err)
// 	require.Equal(t, http.StatusOK, code)

// 	require.Equal(t, 2, len(resp.Responses))
// 	require.Equal(t, "1", resp.Responses[0].ID)
// 	require.Equal(t, 200, resp.Responses[0].Status)
// 	var channelOut api.Channel
// 	err = resp.Responses[0].As(&channelOut)
// 	require.NoError(t, err)
// 	require.Equal(t, keys.ID("kex1fzlrdfy4wlyaturcqkfq92ywj7lft9awtdg70d2yftzhspmc45qsvghhep"), channelOut.ID)
// 	require.Equal(t, int64(1234567890004), channelOut.Timestamp)
// 	require.Equal(t, "2", resp.Responses[1].ID)
// 	require.Equal(t, 200, resp.Responses[1].Status)
// 	err = resp.Responses[1].As(&channelOut)
// 	require.NoError(t, err)
// 	require.Equal(t, keys.ID("kex1tan3x22v8nc6s98gmr9q3zwmy0ngywm4yja0zdylh37e752jj3dsur2s3g"), channelOut.ID)
// 	require.Equal(t, int64(1234567890008), channelOut.Timestamp)
// }
