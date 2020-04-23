package service

import (
	"bufio"
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/urfave/cli"
)

func signCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "sign",
			Usage: "Create a signed message",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "signer, s", Usage: "signer"},
				cli.BoolFlag{Name: "armor, a", Usage: "armored string output"},
				cli.BoolFlag{Name: "detached, d", Usage: "only output signature bytes"},
				cli.StringFlag{Name: "in, i", Usage: "file to read"},
				cli.StringFlag{Name: "out, o", Usage: "file to write (defaults to {in}.sig)"},
			},
			Action: func(c *cli.Context) error {
				if c.String("in") != "" {
					return signFileForCLI(c, client)
				}

				reader := bufio.NewReader(os.Stdin)
				writer := os.Stdout

				client, streamErr := client.ProtoClient().SignStream(context.TODO())
				if streamErr != nil {
					return streamErr
				}
				if err := client.Send(&SignInput{
					Signer:   c.String("signer"),
					Armored:  c.Bool("armor"),
					Detached: c.Bool("detached"),
				}); err != nil {
					return err
				}

				var err error
				go func() {
					_, inErr := readFrom(reader, 1024*1024, func(b []byte) error {
						if len(b) > 0 {
							if err := client.Send(&SignInput{
								Data: b,
							}); err != nil {
								return err
							}
						} else {
							if err := client.CloseSend(); err != nil {
								return err
							}
						}
						return nil
					})
					if inErr != nil {
						err = inErr
					}
				}()

				for {
					resp, recvErr := client.Recv()
					if recvErr != nil {
						if recvErr == io.EOF {
							return err
						}
						return recvErr
					}

					_, writeErr := writer.Write(resp.Data)
					if writeErr != nil {
						return writeErr
					}
				}
			},
		},
	}
}

func signFileForCLI(c *cli.Context, client *Client) error {
	return signFile(client, c.String("signer"), c.Bool("armor"), c.Bool("detached"), c.String("in"), c.String("out"))
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
