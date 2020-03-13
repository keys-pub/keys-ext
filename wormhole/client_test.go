package wormhole_test

import (
	"testing"

	"github.com/keys-pub/keysd/wormhole"
)

func Test(t *testing.T) {
	wormhole.SetLogger(wormhole.NewLogger(wormhole.DebugLevel))

	t.Fatal("testing")
}
