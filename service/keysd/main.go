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
		ServiceName:    "keysd",
		CmdName:        "keys",
		DefaultAppName: "Keys",
		DefaultPort:    22405,
		Description:    "Key management, signing and encryption.",
	}
	service.Run(build)
}
