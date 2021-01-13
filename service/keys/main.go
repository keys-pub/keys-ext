package main

import (
	"github.com/keys-pub/keys-ext/service"
)

// build flags passed from goreleaser
var (
	version = service.VersionDev
	commit  = "snapshot"
	date    = ""
)

func main() {
	build := service.Build{
		Version:        version,
		Commit:         commit,
		Date:           date,
		DefaultAppName: "Keys",
		DefaultPort:    22405,
		ServiceName:    "keysd",
		CmdName:        "keys",
		Description:    "Key management, signing and encryption.",
	}
	service.RunClient(build)
}
