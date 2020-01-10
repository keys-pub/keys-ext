package service

import (
	"bytes"

	"github.com/keys-pub/keys"
)

var alice, _ = keys.NewSignKeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
var bob, _ = keys.NewSignKeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))
var charlie, _ = keys.NewSignKeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x03}, 32)))
var group, _ = keys.NewSignKeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x04}, 32)))
