package main

import (
	"github.com/keys-pub/keys-ext/auth/rpc"
)

// AuthServer exported for plugin.
var AuthServer = rpc.Server{} // nolint

// This is a plugin, so main isn't necessary, but we need it for goreleaser.
func main() {}
