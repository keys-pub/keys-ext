package main

import (
	"log"
	"os"
	"time"

	"github.com/keys-pub/keysd/fido2"
	"github.com/keys-pub/keysd/fido2/cmds"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "fido2"
	app.Version = "1.4.0"
	app.Usage = "Manage FIDO2 devices"

	app.Flags = []cli.Flag{}

	logger := logrus.StandardLogger()
	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339Nano,
	}
	logger.SetFormatter(formatter)

	server := fido2.NewAuthenticatorsServer()

	cliCmds := []cli.Command{
		cmds.Devices(server),
		cmds.DeviceInfo(server),
		cmds.MakeCredential(server),
	}
	// sort.Slice(cliCmds, func(i, j int) bool {
	// 	return cliCmds[i].Name < cliCmds[j].Name
	// })

	app.Commands = cliCmds

	app.Before = func(c *cli.Context) error {
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
