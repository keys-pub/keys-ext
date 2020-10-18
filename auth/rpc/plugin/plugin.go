package main

import (
	"github.com/keys-pub/keys-ext/auth/rpc"
)

// FIDO2Server exported for plugin.
var FIDO2Server = rpc.Server{} // nolint

// This is a plugin, so main isn't necessary, but we need it for goreleaser.
func main() {}
