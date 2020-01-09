package service

import (
	"bytes"
	"fmt"

	"github.com/keys-pub/keys"
)

var alice, _ = keys.NewSignKeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
var bob, _ = keys.NewSignKeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))
var charlie, _ = keys.NewSignKeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x03}, 32)))
var group, _ = keys.NewSignKeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x04}, 32)))

func testPasswordForKey(key *keys.SignKey) string {
	switch key.ID() {
	case "kpe132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqlrnuen":
		return "aaaaaaaaaa"
	case "kpe1syuhwr4g05t4744r23nvxnr7en9cmz53knhr0gja7c84hr7fkw2qrt73l9":
		return "bbbbbbbbbb"
	case "kpe1a4yj333g68pvd6hfqvufqkv4vy54jfe6t33ljd3kc9rpfty8xlgs474npw":
		return "cccccccccc"
	case "kpe1e2f6c9c9rpc8r4nms0rl7rh7syyw3mz9xpt46aexs7fn8k76he7q0lsxqg":
		return "gggggggggg"
	default:
		panic(fmt.Sprintf("unknown test key: %s", key.ID()))
	}
}

func testBackupForKey(key *keys.SignKey) string {
	return seedToBackup(testPasswordForKey(key), key.Seed()[:])
}
