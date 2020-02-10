package service

import (
	"context"
	"io"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func modeFromString(s string) (EncryptMode, error) {
	switch s {
	case "", "encrypt":
		return EncryptV2, nil
	case "signcrypt":
		return SigncryptV1, nil
	default:
		return DefaultEncryptMode, errors.Errorf("invalid mode %s", s)
	}
}

func sealCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:      "encrypt",
			Usage:     "Encrypt",
			ArgsUsage: "<stdin or -in>",
			Flags: []cli.Flag{
				cli.StringSliceFlag{Name: "recipients, r", Usage: "recipients"},
				cli.StringFlag{Name: "signer, s", Usage: "signer (or anonymous if not specified)"},
				cli.BoolFlag{Name: "armor, a", Usage: "armored"},
				cli.StringFlag{Name: "in, i", Usage: "file to read or stdin if not specified"},
				cli.StringFlag{Name: "out, o", Usage: "file to write or stdout if not specified"},
				cli.StringFlag{Name: "mode, m", Usage: "encryption mode: encrypt (default) or signcrypt"},
			},
			Action: func(c *cli.Context) error {
				reader, err := readerFromArgs(c.String("in"))
				if err != nil {
					return err
				}
				writer, err := writerFromArgs(c.String("out"))
				if err != nil {
					return err
				}

				client, err := client.ProtoClient().EncryptStream(context.TODO())
				if err != nil {
					return err
				}
				mode, err := modeFromString(c.String("mode"))
				if err != nil {
					return err
				}

				if err := client.Send(&EncryptStreamInput{
					Recipients: c.StringSlice("recipients"),
					Signer:     c.String("signer"),
					Armored:    c.Bool("armor"),
					Mode:       mode,
				}); err != nil {
					return err
				}

				var readErr error
				go func() {
					_, inErr := readFrom(reader, 1024*1024, func(b []byte) error {
						if len(b) > 0 {
							if err := client.Send(&EncryptStreamInput{Data: b}); err != nil {
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
						readErr = inErr
					}
				}()

				for {
					resp, recvErr := client.Recv()
					if recvErr != nil {
						if recvErr == io.EOF {
							// Return readErr, if set from readStdin above
							return readErr
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
		cli.Command{
			Name:  "decrypt",
			Usage: "Decrypt",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "armor, a", Usage: "armored"},
				cli.StringFlag{Name: "in, i", Usage: "file to read or stdin if not specified"},
				cli.StringFlag{Name: "out, o", Usage: "file to write or stdout if not specified"},
				cli.StringFlag{Name: "mode, m", Usage: "encryption mode: encrypt (default) or signcrypt"},
			},
			ArgsUsage: "<stdin or -in>",
			Action: func(c *cli.Context) error {
				reader, err := readerFromArgs(c.String("in"))
				if err != nil {
					return err
				}
				writer, err := writerFromArgs(c.String("out"))
				if err != nil {
					return err
				}
				mode, err := modeFromString(c.String("mode"))
				if err != nil {
					return err
				}

				client, err := NewDecryptStreamClient(context.TODO(), client.ProtoClient(), c.Bool("armor"), mode)
				if err != nil {
					return err
				}
				var openErr error
				go func() {
					_, inErr := readFrom(reader, 1024*1024, func(b []byte) error {
						if len(b) > 0 {
							if err := client.Send(&DecryptStreamInput{Data: b}); err != nil {
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
						openErr = inErr
					}
				}()

				wgOpen := sync.WaitGroup{}
				wgOpen.Add(1)
				go func() {
					for {
						resp, recvErr := client.Recv()
						if recvErr != nil {
							if recvErr == io.EOF {
								break
							}
							openErr = recvErr
							break
						}
						if resp.Signer != nil {
							fmtKey(os.Stdout, resp.Signer, "verified ")
						}
						if len(resp.Data) == 0 {
							break
						}

						if err := writeAll(writer, resp.Data); err != nil {
							openErr = err
							break
						}
					}
					wgOpen.Done()
				}()
				wgOpen.Wait()

				return openErr
			},
		},
	}
}
