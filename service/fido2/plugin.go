package main

import (
	"github.com/keys-pub/keysd/fido2/authenticators"
)

// AuthenticatorsServer exported for plugin.
var AuthenticatorsServer = authenticators.Server{}

// This is a plugin, so main isn't necessary, but we need it for goreleaser.
func main() {}
