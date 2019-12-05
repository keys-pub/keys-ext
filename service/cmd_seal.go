package service

import (
	"context"
	"io"
	"sync"

	"github.com/urfave/cli"
)

// sealCommands ...
func sealCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:      "encrypt",
			Usage:     "Encrypt",
			Aliases:   []string{"seal"},
			ArgsUsage: "<stdin or -in>",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "recipients, r", Usage: "recipients kids, comma-seperated"},
				cli.StringFlag{Name: "sender, s", Usage: "sender kid (defaults to current key)"},
				cli.BoolFlag{Name: "armor, a", Usage: "armored"},
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

				client, streamErr := client.ProtoClient().EncryptStream(context.TODO())
				if streamErr != nil {
					return streamErr
				}
				if err := client.Send(&EncryptStreamInput{
					Recipients: c.String("recipients"),
					Sender:     c.String("sender"),
					Armored:    c.Bool("armor"),
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
			Name:    "decrypt",
			Usage:   "Decrypt",
			Aliases: []string{"open"},
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "armor, a", Usage: "armored"},
				cli.StringFlag{Name: "in, i", Usage: "file to read or stdin if not specified"},
				cli.StringFlag{Name: "out, o", Usage: "file to write or stdout if not specified"},
			},
			ArgsUsage: "<stdin or -in>",
			Action: func(c *cli.Context) error {
				reader, readerErr := readerFromArgs(c.String("in"))
				if readerErr != nil {
					return readerErr
				}
				writer, writerErr := writerFromArgs(c.String("out"))
				if writerErr != nil {
					return writerErr
				}

				client, clientErr := NewDecryptStreamClient(context.TODO(), client.ProtoClient(), c.Bool("armor"))
				if clientErr != nil {
					return clientErr
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

func writeAll(writer io.Writer, b []byte) error {
	n, writeErr := writer.Write(b)
	if writeErr != nil {
		return writeErr
	}
	if n != len(b) {
		return io.ErrShortWrite
	}
	return nil
}
