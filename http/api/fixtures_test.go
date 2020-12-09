package api_test

import (
	"bytes"

	"github.com/keys-pub/keys"
)

var alice = keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
var bob = keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))
var charlie = keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x03}, 32)))

var channel = keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0xef}, 32)))
