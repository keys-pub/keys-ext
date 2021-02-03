package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func signCommands(client *Client) []cli.Command {
	return []cli.Command{
		{
			Name:  "sign",
			Usage: "Sign",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "signer, s", Usage: "signer"},
				cli.StringFlag{Name: "in, i", Usage: "file to read"},
				cli.StringFlag{Name: "out, o", Usage: "file to write, defaults to <in>.sig (detached) or <in>.signed (attached)"},

				cli.BoolFlag{Name: "attached, t", Usage: "output attached signature (.signed)"},
				cli.BoolFlag{Name: "detached, d", Usage: "output detached signature (.sig)"},
				cli.BoolFlag{Name: "armor, a", Usage: "armored"},
				cli.BoolFlag{Name: "binary, b", Usage: "binary"},
			},
			Action: func(c *cli.Context) error {
				mode, err := parseMode(c)
				if err != nil {
					return err
				}

				if c.String("in") != "" {
					return signFileForCLI(c, client, mode)
				}

				return signStdin(c, client, mode)
			},
		},
	}
}

func signStdin(c *cli.Context, client *Client, mode signMode) error {
	reader := bufio.NewReader(os.Stdin)
	writer := os.Stdout
	detached := mode.isDetached(stdIn)

	signClient, streamErr := client.RPCClient().SignStream(context.TODO())
	if streamErr != nil {
		return streamErr
	}
	if err := signClient.Send(&SignInput{
		Signer:   c.String("signer"),
		Armored:  mode.isArmored(stdIn, detached),
		Detached: detached,
	}); err != nil {
		return err
	}

	var readErr error
	go func() {
		_, inErr := readFrom(reader, 1024*1024, func(b []byte) error {
			if len(b) > 0 {
				if err := signClient.Send(&SignInput{
					Data: b,
				}); err != nil {
					return err
				}
			} else {
				if err := signClient.CloseSend(); err != nil {
					return err
				}
			}
			return nil
		})
		if inErr != nil {
			readErr = inErr
		}
	}()

	for {
		resp, recvErr := signClient.Recv()
		if recvErr != nil {
			if recvErr == io.EOF {
				return readErr
			}
			return recvErr
		}

		_, writeErr := writer.Write(resp.Data)
		if writeErr != nil {
			return writeErr
		}
	}
}

func signFileForCLI(c *cli.Context, client *Client, mode signMode) error {
	detached := mode.isDetached(fileIn)
	return signFile(client, c.String("signer"), mode.isArmored(fileIn, detached), detached, c.String("in"), c.String("out"))
}

func signFile(client *Client, signer string, armored bool, detached bool, in string, out string) error {
	in, err := filepath.Abs(in)
	if err != nil {
		return err
	}
	if out != "" {
		out, err = filepath.Abs(out)
		if err != nil {
			return err
		}
	}

	signClient, err := client.RPCClient().SignFile(context.TODO())
	if err != nil {
		return err
	}

	if err := signClient.Send(&SignFileInput{
		Signer:   signer,
		Armored:  armored,
		Detached: detached,
		In:       in,
		Out:      out,
	}); err != nil {
		return err
	}

	resp, recvErr := signClient.Recv()
	if recvErr != nil {
		// if recvErr == io.EOF {
		// 	break
		// }
		return recvErr
	}
	// if err := encryptClient.CloseSend(); err != nil {
	// 	return err
	// }

	if resp.Out != "" {
		fmt.Fprintf(client.out, "out: %s\n", resp.Out)
	}

	return nil
}

type option string

const (
	defaultOption option = "default"
	trueOption    option = "true"
	falseOption   option = "false"
)

type signMode struct {
	armored  option
	detached option
}

type inputType string

const (
	stdIn  inputType = "stdin"
	fileIn inputType = "file"
)

func (m signMode) isArmored(it inputType, detached bool) bool {
	if m.armored == defaultOption {
		if detached || it == stdIn {
			return true
		}
		return false
	}
	return m.armored == trueOption
}

func (m signMode) isDetached(it inputType) bool {
	if m.detached == defaultOption {
		switch it {
		case fileIn:
			return true
		default:
			return false
		}
	}
	return m.detached == trueOption
}

func parseMode(c *cli.Context) (signMode, error) {
	mode := signMode{
		armored:  defaultOption,
		detached: defaultOption,
	}

	if c.Bool("attached") {
		mode.detached = falseOption
	}
	if c.Bool("detached") {
		mode.detached = trueOption
	}
	if c.Bool("armor") {
		mode.armored = trueOption
	}
	if c.Bool("binary") {
		mode.armored = falseOption
	}

	// Check conflicts
	if c.Bool("attached") && c.Bool("detached") {
		return mode, errors.Errorf("conflicting attached and detached options")
	}
	if c.Bool("armor") && c.Bool("binary") {
		return mode, errors.Errorf("conflicting armor and binary options")
	}

	return mode, nil
}
