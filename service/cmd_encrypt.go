package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func encryptCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "encrypt",
			Usage: "Encrypt",
			Flags: []cli.Flag{
				cli.StringSliceFlag{Name: "recipient, r", Usage: "recipients"},
				cli.StringFlag{Name: "sender, signer, s", Usage: "signer (or anonymous)"},
				cli.BoolFlag{Name: "armor, a", Usage: "armored"},
				cli.StringFlag{Name: "in, i", Usage: "file to read"},
				cli.StringFlag{Name: "out, o", Usage: "file to write, defaults to <in>.enc"},
				cli.StringFlag{Name: "mode, m", Usage: "encryption mode: signcrypt, encrypt or default (signcrypt if signing, encrypt otherwise)"},
				cli.BoolFlag{Hidden: true, Name: "no-signer-recipient", Usage: "don't add signer to recipients"},
			},
			Action: func(c *cli.Context) error {
				if c.String("in") == "" && c.String("out") != "" {
					return errors.Errorf("-out option is unsupported without -in")
				}

				mode, err := encryptModeFromString(c.String("mode"))
				if err != nil {
					return err
				}
				options := &EncryptOptions{
					Mode:              mode,
					Armored:           c.Bool("armor"),
					NoSenderRecipient: c.Bool("no-sender-recipient"),
				}

				if c.String("in") != "" {
					return encryptFileForCLI(c, client, options)
				}

				reader := bufio.NewReader(os.Stdin)
				writer := os.Stdout

				encryptClient, err := client.KeysClient().EncryptStream(context.TODO())
				if err != nil {
					return err
				}

				if err := encryptClient.Send(&EncryptInput{
					Recipients: c.StringSlice("recipient"),
					Sender:     c.String("sender"),
					Options:    options,
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
				cli.StringFlag{Name: "in, i", Usage: "file to read"},
				cli.StringFlag{Name: "out, o", Usage: "file to write"},
			},
			Action: func(c *cli.Context) error {
				if c.String("in") != "" {
					dec, err := decryptFileForCLI(c, client)
					if err != nil {
						return err
					}
					if dec.Sender != nil {
						fmtVerifiedEncrypt(client.out, dec.Sender, dec.Mode)
					}
					if dec.Out != c.String("out") {
						fmt.Fprintf(client.out, "out: %s\n", dec.Out)
					}
					return nil
				}
				reader := bufio.NewReader(os.Stdin)
				writer := os.Stdout

				decryptClient, err := NewDecryptStreamClient(context.TODO(), client.KeysClient())
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
					showSender := true
					for {
						resp, recvErr := decryptClient.Recv()
						if recvErr != nil {
							if recvErr == io.EOF {
								break
							}
							openErr = recvErr
							break
						}
						if showSender && resp.Sender != nil {
							fmtVerifiedEncrypt(client.out, resp.Sender, resp.Mode)
							showSender = false
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

func encryptFileForCLI(c *cli.Context, client *Client, options *EncryptOptions) error {
	return encryptFile(client, c.String("in"), c.String("out"), c.StringSlice("recipient"), c.String("sender"), options)
}

func encryptFile(client *Client, in string, out string, recipients []string, sender string, options *EncryptOptions) error {
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

	encryptClient, err := client.KeysClient().EncryptFile(context.TODO())
	if err != nil {
		return err
	}

	if err := encryptClient.Send(&EncryptFileInput{
		In:         in,
		Out:        out,
		Recipients: recipients,
		Sender:     sender,
		Options:    options,
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
	return decryptFile(client, c.String("in"), c.String("out"))
}

func decryptFile(client *Client, in string, out string) (*DecryptFileOutput, error) {
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

	decryptClient, err := client.KeysClient().DecryptFile(context.TODO())
	if err != nil {
		return nil, err
	}

	if err := decryptClient.Send(&DecryptFileInput{
		In:  in,
		Out: out,
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

func encryptModeFromString(s string) (EncryptMode, error) {
	switch s {
	case "", "default":
		return DefaultEncrypt, nil
	case "encrypt":
		return SaltpackEncrypt, nil
	case "signcrypt":
		return SaltpackSigncrypt, nil
	default:
		return DefaultEncrypt, errors.Errorf("invalid mode %q", s)
	}
}
