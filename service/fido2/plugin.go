package main

import (
	"github.com/keys-pub/keysd/fido2/libfido2"
)

// AuthenticatorsServer exported for plugin.
var AuthenticatorsServer = libfido2.Server{}

// This is a plugin, so main isn't necessary, but we need it for goreleaser.
func main() {}
