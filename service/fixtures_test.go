package service

import (
	"bytes"

	"github.com/keys-pub/keys"
)

var alice = keys.NewEd25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
var bob = keys.NewEd25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))
var charlie = keys.NewEd25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x03}, 32)))
var group = keys.NewEd25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x04}, 32)))
