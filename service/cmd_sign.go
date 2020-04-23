package service

import (
	"bufio"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func signCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "sign",
			Usage: "Sign",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "signer, s", Usage: "signer"},
				cli.StringFlag{Name: "mode, m", Value: "", Usage: "override defaults, armor | binary | attached | detached | armor,attached ..."},
				cli.StringFlag{Name: "in, i", Usage: "file to read"},
				cli.StringFlag{Name: "out, o", Usage: "file to write, defaults to {in}.sig (detached) or {in}.signed (attached)"},
			},
			Action: func(c *cli.Context) error {
				mode, err := parseMode(c.String("mode"), false)
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

	signClient, streamErr := client.ProtoClient().SignStream(context.TODO())
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

	signClient, err := client.ProtoClient().SignFile(context.TODO())
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

	_, recvErr := signClient.Recv()
	if recvErr != nil {
		// if recvErr == io.EOF {
		// 	break
		// }
		return recvErr
	}
	// if err := encryptClient.CloseSend(); err != nil {
	// 	return err
	// }

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

func parseMode(s string, sig bool) (signMode, error) {
	mode := signMode{
		armored:  defaultOption,
		detached: defaultOption,
	}

	if sig {
		mode.detached = trueOption
	}

	if s == "" {
		return mode, nil
	}

	strs := strings.Split(s, ",")
	for _, str := range strs {
		switch str {
		case "armor":
			mode.armored = trueOption
		case "binary":
			mode.armored = falseOption
		case "detached":
			mode.detached = trueOption
		case "attached":
			mode.detached = falseOption
		default:
			return mode, errors.Errorf("invalid mode %s", str)
		}
	}

	return mode, nil
}
