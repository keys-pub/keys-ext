package service

import (
	"context"
	"io"
	"os"
	"path/filepath"
	strings "strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func checkSigner(signer *Key, expected string) error {
	if signer == nil {
		return nil
	}

	if strings.Contains(expected, "@") {
		if signer.User == nil {
			return errors.Errorf("invalid signer, expected %s, was %s", expected, signer.ID)
		}
		if signer.User.Label != expected {
			return errors.Errorf("invalid signer, expected %s, was %s", expected, signer.User.Label)
		}
		return nil
	}

	if signer.ID != expected {
		return errors.Errorf("invalid signer, expected %s, was %s", expected, signer.ID)
	}

	return nil
}

func signCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:      "sign",
			Usage:     "Create a signed message",
			ArgsUsage: "<stdin or -in>",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "signer, s", Usage: "signer"},
				cli.BoolFlag{Name: "armor, a", Usage: "armored string output"},
				cli.BoolFlag{Name: "detached, d", Usage: "only output signature bytes"},
				cli.StringFlag{Name: "in, i", Usage: "file to read or stdin if not specified"},
				cli.StringFlag{Name: "out, o", Usage: "file to write or stdout if not specified"},
			},
			Action: func(c *cli.Context) error {
				if c.String("in") != "" && c.String("out") != "" {
					return signFileForCLI(c, client)
				}

				reader, readerErr := readerFromArgs(c.String("in"))
				if readerErr != nil {
					return readerErr
				}
				writer, writerErr := writerFromArgs(c.String("out"))
				if writerErr != nil {
					return writerErr
				}

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
		cli.Command{
			Name:      "verify",
			Usage:     "Verify a signed message",
			ArgsUsage: "<message>",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "signer, s", Usage: "expected signer"},
				cli.BoolFlag{Name: "armor, a"},
				cli.StringFlag{Name: "in, i", Usage: "file to read or stdin if not specified"},
				cli.StringFlag{Name: "out, o", Usage: "file to write or stdout if not specified"},
			},
			Action: func(c *cli.Context) error {
				signer := c.String("signer")

				if c.String("in") != "" && c.String("out") != "" {
					verified, err := verifyFileForCLI(c, client)
					if err != nil {
						return err
					}
					if signer != "" {
						if err := checkSigner(verified, signer); err != nil {
							return err
						}
					} else if verified != nil {
						fmtKey(os.Stdout, verified, "verified ")
					}
					return nil
				}

				reader, readerErr := readerFromArgs(c.String("in"))
				if readerErr != nil {
					return readerErr
				}
				writer, writerErr := writerFromArgs(c.String("out"))
				if writerErr != nil {
					return writerErr
				}

				client, clientErr := NewVerifyStreamClient(context.TODO(), client.ProtoClient(), c.Bool("armor"))
				if clientErr != nil {
					return clientErr
				}
				var outErr error
				go func() {
					_, inErr := readFrom(reader, 1024*1024, func(b []byte) error {
						if len(b) > 0 {
							if err := client.Send(&VerifyInput{Data: b}); err != nil {
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
						outErr = inErr
					}
				}()

				wg := sync.WaitGroup{}
				wg.Add(1)
				go func() {
					for {
						resp, recvErr := client.Recv()
						if recvErr != nil {
							if recvErr == io.EOF {
								break
							}
							outErr = recvErr
							break
						}
						if signer != "" {
							if err := checkSigner(resp.Signer, signer); err != nil {
								outErr = err
								break
							}
						} else if resp.Signer != nil {
							fmtKey(os.Stdout, resp.Signer, "verified ")
						}

						if len(resp.Data) == 0 {
							break
						}

						_, writeErr := writer.Write(resp.Data)
						if writeErr != nil {
							outErr = writeErr
							break
						}
					}
					wg.Done()
				}()
				wg.Wait()

				return outErr
			},
		},
	}
}

func signFileForCLI(c *cli.Context, client *Client) error {
	return signFile(client, c.String("signer"), c.Bool("armored"), c.Bool("detached"), c.String("in"), c.String("out"))
}

func signFile(client *Client, signer string, armored bool, detached bool, in string, out string) error {
	in, err := filepath.Abs(in)
	if err != nil {
		return err
	}
	out, err = filepath.Abs(out)
	if err != nil {
		return err
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

func verifyFileForCLI(c *cli.Context, client *Client) (*Key, error) {
	return verifyFile(client, c.Bool("armored"), c.String("in"), c.String("out"))
}

func verifyFile(client *Client, armored bool, in string, out string) (*Key, error) {
	in, err := filepath.Abs(in)
	if err != nil {
		return nil, err
	}
	out, err = filepath.Abs(out)
	if err != nil {
		return nil, err
	}

	verifyClient, err := client.ProtoClient().VerifyFile(context.TODO())
	if err != nil {
		return nil, err
	}

	if err := verifyClient.Send(&VerifyFileInput{
		Armored: armored,
		In:      in,
		Out:     out,
	}); err != nil {
		return nil, err
	}

	resp, recvErr := verifyClient.Recv()
	if recvErr != nil {
		// if recvErr == io.EOF {
		// 	break
		// }
		return nil, recvErr
	}
	// if err := encryptClient.CloseSend(); err != nil {
	// 	return err
	// }

	return resp.Signer, nil
}
