package service

import (
	"context"
	"io"
	"os"
	"path/filepath"
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

func encryptCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:      "encrypt",
			Usage:     "Encrypt",
			ArgsUsage: "stdin or -in",
			Flags: []cli.Flag{
				cli.StringSliceFlag{Name: "recipient, r", Usage: "recipients"},
				cli.StringFlag{Name: "sender, s", Usage: "sender (or anonymous if not specified)"},
				cli.BoolFlag{Name: "armor, a", Usage: "armored"},
				cli.StringFlag{Name: "in, i", Usage: "file to read or stdin if not specified"},
				cli.StringFlag{Name: "out, o", Usage: "file to write or stdout if not specified"},
				cli.StringFlag{Name: "mode, m", Usage: "encryption mode: encrypt (default) or signcrypt"},
			},
			Action: func(c *cli.Context) error {
				if c.String("in") != "" && c.String("out") != "" {
					return encryptFileForCLI(c, client)
				}

				mode, err := modeFromString(c.String("mode"))
				if err != nil {
					return err
				}
				reader, err := readerFromArgs(c.String("in"))
				if err != nil {
					return err
				}
				writer, err := writerFromArgs(c.String("out"))
				if err != nil {
					return err
				}

				encryptClient, err := client.ProtoClient().EncryptStream(context.TODO())
				if err != nil {
					return err
				}

				if err := encryptClient.Send(&EncryptInput{
					Recipients: c.StringSlice("recipient"),
					Sender:     c.String("sender"),
					Armored:    c.Bool("armor"),
					Mode:       mode,
				}); err != nil {
					return err
				}

				var readErr error
				go func() {
					_, inErr := readFrom(reader, 1024*1024, func(b []byte) error {
						if len(b) > 0 {
							if err := encryptClient.Send(&EncryptInput{Data: b}); err != nil {
								return err
							}
						} else {
							if err := encryptClient.CloseSend(); err != nil {
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
					resp, recvErr := encryptClient.Recv()
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
			ArgsUsage: "stdin or -in",
			Action: func(c *cli.Context) error {
				if c.String("in") != "" && c.String("out") != "" {
					dec, err := decryptFileForCLI(c, client)
					if err != nil {
						return err
					}
					if dec.Sender != nil {
						fmtKey(os.Stdout, dec.Sender, "verified ")
					}
					return nil
				}
				mode, err := modeFromString(c.String("mode"))
				if err != nil {
					return err
				}
				reader, err := readerFromArgs(c.String("in"))
				if err != nil {
					return err
				}
				writer, err := writerFromArgs(c.String("out"))
				if err != nil {
					return err
				}

				decryptClient, err := NewDecryptStreamClient(context.TODO(), client.ProtoClient(), c.Bool("armor"), mode)
				if err != nil {
					return err
				}
				var openErr error
				go func() {
					_, inErr := readFrom(reader, 1024*1024, func(b []byte) error {
						if len(b) > 0 {
							if err := decryptClient.Send(&DecryptInput{Data: b}); err != nil {
								return err
							}
						} else {
							if err := decryptClient.CloseSend(); err != nil {
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
						resp, recvErr := decryptClient.Recv()
						if recvErr != nil {
							if recvErr == io.EOF {
								break
							}
							openErr = recvErr
							break
						}
						if resp.Sender != nil {
							fmtKey(os.Stdout, resp.Sender, "verified ")
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

func encryptFileForCLI(c *cli.Context, client *Client) error {
	mode, err := modeFromString(c.String("mode"))
	if err != nil {
		return err
	}
	return encryptFile(client, c.StringSlice("recipient"), c.String("sender"), c.Bool("armored"), mode, c.String("in"), c.String("out"))
}

func encryptFile(client *Client, recipients []string, sender string, armored bool, mode EncryptMode, in string, out string) error {
	in, err := filepath.Abs(in)
	if err != nil {
		return err
	}
	out, err = filepath.Abs(out)
	if err != nil {
		return err
	}

	encryptClient, err := client.ProtoClient().EncryptFile(context.TODO())
	if err != nil {
		return err
	}

	if err := encryptClient.Send(&EncryptFileInput{
		Recipients: recipients,
		Sender:     sender,
		Armored:    armored,
		Mode:       mode,
		In:         in,
		Out:        out,
	}); err != nil {
		return err
	}

	_, recvErr := encryptClient.Recv()
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

func decryptFileForCLI(c *cli.Context, client *Client) (*DecryptFileOutput, error) {
	mode, err := modeFromString(c.String("mode"))
	if err != nil {
		return nil, err
	}
	return decryptFile(client, c.Bool("armored"), mode, c.String("in"), c.String("out"))
}

func decryptFile(client *Client, armored bool, mode EncryptMode, in string, out string) (*DecryptFileOutput, error) {
	if in == "" {
		return nil, errors.Errorf("in not specified")
	}
	in, err := filepath.Abs(in)
	if err != nil {
		return nil, err
	}
	if out != "" {
		out, err = filepath.Abs(out)
		if err != nil {
			return nil, err
		}
	}

	decryptClient, err := client.ProtoClient().DecryptFile(context.TODO())
	if err != nil {
		return nil, err
	}

	if err := decryptClient.Send(&DecryptFileInput{
		Armored: armored,
		Mode:    mode,
		In:      in,
		Out:     out,
	}); err != nil {
		return nil, err
	}

	resp, recvErr := decryptClient.Recv()
	if recvErr != nil {
		// if recvErr == io.EOF {
		// 	break
		// }
		return nil, recvErr
	}
	// if err := decryptClient.CloseSend(); err != nil {
	// 	return err
	// }

	return resp, nil
}
