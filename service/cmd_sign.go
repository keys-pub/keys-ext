package service

import (
	"context"
	"io"
	"sync"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func signCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:      "sign",
			Usage:     "Create a signed message",
			ArgsUsage: "<stdin or -in>",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "kid, k", Usage: "kid to sign (defaults to current key)"},
				cli.BoolFlag{Name: "armor, a", Usage: "armored string output"},
				cli.BoolFlag{Name: "detached, d", Usage: "only output signature bytes"},
				cli.StringFlag{Name: "in, i", Usage: "file to read or stdin if not specified"},
				cli.StringFlag{Name: "out, o", Usage: "file to write or stdout if not specified"},
			},
			Action: func(c *cli.Context) error {
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
				if err := client.Send(&SignStreamInput{
					KID:      c.String("kid"),
					Armored:  c.Bool("armor"),
					Detached: c.Bool("detached"),
				}); err != nil {
					return err
				}

				var err error
				go func() {
					_, inErr := readFrom(reader, 1024*1024, func(b []byte) error {
						if len(b) > 0 {
							if err := client.Send(&SignStreamInput{
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
				cli.StringFlag{Name: "kid, k"},
				cli.BoolFlag{Name: "armor, a"},
				cli.StringFlag{Name: "in, i", Usage: "file to read or stdin if not specified"},
				cli.StringFlag{Name: "out, o", Usage: "file to write or stdout if not specified"},
			},
			Action: func(c *cli.Context) error {
				kid := c.String("kid")
				if kid == "" {
					return errors.Errorf("kid not specified")
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
				var err error
				go func() {
					_, inErr := readFrom(reader, 1024*1024, func(b []byte) error {
						if len(b) > 0 {
							if err := client.Send(&VerifyStreamInput{Data: b}); err != nil {
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

				wg := sync.WaitGroup{}
				wg.Add(1)
				go func() {
					for {
						resp, recvErr := client.Recv()
						if recvErr != nil {
							if recvErr == io.EOF {
								break
							}
							err = recvErr
							break
						}
						if (kid == "" && resp.Signer == nil) || (resp.Signer != nil && resp.Signer.ID == kid) {
							// OK
						} else {
							err = errors.Errorf("not signed by the specified kid, expected %s, was %s", kid, resp.Signer.ID)
							break
						}

						if len(resp.Data) == 0 {
							break
						}

						_, writeErr := writer.Write(resp.Data)
						if writeErr != nil {
							err = writeErr
							break
						}
					}
					wg.Done()
				}()
				wg.Wait()

				return err
			},
		},
	}
}
