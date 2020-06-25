package client_test

import (
	"testing"

	"github.com/keys-pub/keys-ext/http/client"
	"github.com/stretchr/testify/require"
)

func TestEventchain(t *testing.T) {
	var err error

	event3 := client.NewEvent("/col1/key3", []byte("test3"), nil)
	event4a := client.NewEvent("/col1/key4", []byte("test4.1"), event3)
	event4b := client.NewEvent("/col1/key4", []byte("test4.2"), event4a)
	event5 := client.NewEvent("/col1/key5", []byte("test5"), event4b)
	events := []*client.Event{event3, event4a, event4b, event5}

	err = client.CheckEventchain(events)
	require.NoError(t, err)

	eventsBadOrder := []*client.Event{event4a, event3, event4b, event5}
	err = client.CheckEventchain(eventsBadOrder)
	require.EqualError(t, err, "previous event hash not found")
}
