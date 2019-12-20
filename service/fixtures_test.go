package service

import (
	"encoding/hex"

	"github.com/keys-pub/keys"
)

var aliceSeed, _ = hex.DecodeString("d7967d7fc1ed2e09ec4723d7a5eda3f604b4914292cee80f2412005918e2626d")
var alice, _ = keys.NewKey(keys.Bytes32(aliceSeed))
var bobSeed, _ = hex.DecodeString("290cdb738a7def8b3f9368a7ee112297027eb03e7ab77c07d1087ab15c81cd5e")
var bob, _ = keys.NewKey(keys.Bytes32(bobSeed))
var charlieSeed, _ = hex.DecodeString("ba50880c0f969818b216ded861cbe8b78ce4050e4b89a7774cd7a3b106e8c1fb")
var charlie, _ = keys.NewKey(keys.Bytes32(charlieSeed))
var groupSeed, _ = hex.DecodeString("3059dbcfbc1efb47f71fef786d8efa102dc61f96c6f6243987a0969aa5d6d78f")
var group, _ = keys.NewKey(keys.Bytes32(groupSeed))

func testPasswordForKey(key keys.Key) string {
	switch key.ID() {
	case "a6MtPHR36F9wG5orC8bhm8iPCE2xrXK41iZLwPZcLzqo":
		return "aaaaaaaaaa"
	case "bDM13g2wsoBE8WN2jrPdLRHg2LFgNt2ZrLcP2bG4iuNi":
		return "bbbbbbbbbb"
	case "cBYSYNgt45ZULLrVAseoFnmCt87mycnqDF5psywZ53VB":
		return "cccccccccc"
	case "gqPhYydcdbTzHUdqVrrqBnnAJK9tv3gYbrPKPBynjciM":
		return "gggggggggg"
	default:
		panic("unknown test key")
	}
}

func testBackupForKey(key keys.Key) string {
	return seedToBackup(testPasswordForKey(key), key.Seed()[:])
}
