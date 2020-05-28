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
		Version: version,
		Commit:  commit,
		Date:    date,
	}
	service.RunClient(build)
}
