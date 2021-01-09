package service

import (
	"bufio"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// TODO: do we report verified signer only on first or every response, what if a later signed chunk is invalid?

func verifyCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "verify",
			Usage: "Verify",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "signer, s", Usage: "expected signer"},
				cli.StringFlag{Name: "sig, x", Usage: "signature file (if detached)"},
				cli.StringFlag{Name: "in, i", Usage: "file to read"},
				cli.StringFlag{Name: "out, o", Usage: "file to write (if attached), defaults to {in} without .signed"},
			},
			Action: func(c *cli.Context) error {
				logger.Debugf("Verify (cmd)")

				// TODO: Error if out path already exists

				signer := c.String("signer")

				if c.String("in") != "" {
					if _, err := verifyFileForCLI(c, client, signer); err != nil {
						return err
					}
					return nil
				}

				reader := bufio.NewReader(os.Stdin)
				writer := os.Stdout

				sigFile := c.String("sig")
				if sigFile != "" {
					if err := verifyDetachedStream(client, reader, sigFile, signer); err != nil {
						return err
					}
					return nil
				}

				logger.Debugf("Verify stream (cmd)")
				verifyClient, err := NewVerifyStreamClient(context.TODO(), client.RPCClient())
				if err != nil {
					return err
				}
				var outErr error
				go func() {
					_, inErr := readFrom(reader, 1024*1024, func(b []byte) error {
						if len(b) > 0 {
							if err := verifyClient.Send(&VerifyInput{Data: b}); err != nil {
								return err
							}
						} else {
							if err := verifyClient.CloseSend(); err != nil {
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
				checked := false
				go func() {
					for {
						resp, recvErr := verifyClient.Recv()
						if recvErr != nil {
							if recvErr == io.EOF {
								break
							}
							outErr = recvErr
							break
						}
						// TODO: Send expected signer in request
						if !checked {
							if err := checkSigner(os.Stderr, resp.Signer, signer); err != nil {
								outErr = err
								break
							}
							checked = true
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

func verifyFileForCLI(c *cli.Context, client *Client, signer string) (string, error) {
	sigFile := c.String("sig")
	in := c.String("in")

	if sigFile != "" {
		if err := verifyDetachedFile(client, in, sigFile, signer); err != nil {
			return "", err
		}
		return "", nil
	}

	return verifyFile(client, in, c.String("out"), signer)
}

func verifyFile(client *Client, in string, out string, signer string) (string, error) {
	logger.Debugf("Verify file (cmd) in=%s, out=%s", in, out)
	if in == "" {
		return "", errors.Errorf("in not specified")
	}
	in, err := filepath.Abs(in)
	if err != nil {
		return "", err
	}
	if out != "" {
		out, err = filepath.Abs(out)
		if err != nil {
			return "", err
		}
	}

	verifyClient, err := client.RPCClient().VerifyFile(context.TODO())
	if err != nil {
		return "", err
	}
	if err := verifyClient.Send(&VerifyFileInput{
		In:  in,
		Out: out,
	}); err != nil {
		return "", err
	}

	// resp, recvErr := verifyClient.CloseAndRecv()
	// if recvErr != nil {
	// 	return "", recvErr
	// }
	resp, recvErr := verifyClient.Recv()
	if recvErr != nil {
		// if recvErr == io.EOF {
		// 	break
		// }
		return "", recvErr
	}
	// if err := encryptClient.CloseSend(); err != nil {
	// 	return err
	// }

	if err := checkSigner(os.Stderr, resp.Signer, signer); err != nil {
		return "", err
	}

	return resp.Out, nil
}

func verifyDetachedFile(client *Client, in string, sigFile string, signer string) error {
	logger.Debugf("Verify detached file (cmd) in=%s, sig=%s", in, sigFile)
	if in == "" {
		return errors.Errorf("in not specified")
	}
	in, err := filepath.Abs(in)
	if err != nil {
		return err
	}

	sig, err := ioutil.ReadFile(sigFile) // #nosec
	if err != nil {
		return err
	}

	verifyClient, err := client.RPCClient().VerifyDetachedFile(context.TODO())
	if err != nil {
		return err
	}

	if err := verifyClient.Send(&VerifyDetachedFileInput{
		In:  in,
		Sig: sig,
	}); err != nil {
		return err
	}

	resp, recvErr := verifyClient.CloseAndRecv()
	if recvErr != nil {
		return recvErr
	}

	if err := checkSigner(os.Stderr, resp.Signer, signer); err != nil {
		return err
	}

	return nil
}

func verifyDetachedStream(client *Client, reader io.Reader, sigFile string, signer string) error {
	logger.Debugf("Verify detached stream (cmd) sig=%s", sigFile)
	sig, err := ioutil.ReadFile(sigFile) // #nosec
	if err != nil {
		return err
	}

	verifyClient, err := client.RPCClient().VerifyDetachedStream(context.TODO())
	if err != nil {
		return err
	}

	sentSig := false
	var resp *VerifyDetachedResponse

	_, inErr := readFrom(reader, 1024*1024, func(b []byte) error {
		if len(b) > 0 {
			var req *VerifyDetachedInput
			if !sentSig {
				req = &VerifyDetachedInput{
					Sig:  sig,
					Data: b,
				}
				sentSig = true
			} else {
				req = &VerifyDetachedInput{
					Data: b,
				}
			}
			if err := verifyClient.Send(req); err != nil {
				return err
			}
		} else {
			recv, err := verifyClient.CloseAndRecv()
			if err != nil {
				return err
			}
			resp = recv

		}
		return nil
	})
	if inErr != nil {
		return inErr
	}

	if resp == nil {
		return errors.Errorf("no response")
	}

	// TODO: Send expected signer in request
	if err := checkSigner(os.Stderr, resp.Signer, signer); err != nil {
		return err
	}

	return nil
}

func checkSigner(out io.Writer, signer *Key, expected string) error {
	if signer == nil {
		return errors.Errorf("no signer")
	}

	if expected == "" {
		if out != nil {
			fmtVerified(out, signer)
		}
		return nil
	}

	if strings.Contains(expected, "@") {
		if signer.User == nil {
			return errors.Errorf("invalid signer, expected %s, was %s", expected, signer.ID)
		}
		if signer.User.ID != expected {
			return errors.Errorf("invalid signer, expected %s, was %s", expected, signer.User.ID)
		}
		return nil
	}

	if signer.ID != expected {
		return errors.Errorf("invalid signer, expected %s, was %s", expected, signer.ID)
	}

	return nil
}
