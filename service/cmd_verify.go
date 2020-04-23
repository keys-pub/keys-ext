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

func checkSigner(signer *Key, expected string) error {
	if signer == nil {
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

func verifyCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:      "verify",
			Usage:     "Verify a signed message",
			ArgsUsage: "message",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "signer, s", Usage: "expected signer"},
				cli.BoolFlag{Name: "armor, a"},
				cli.StringFlag{Name: "sig, x", Usage: "signature file (if detached)"},
				cli.StringFlag{Name: "in, i", Usage: "file to read"},
				cli.StringFlag{Name: "out, o", Usage: "file to write (defaults to {in} without .sig)"},
			},
			Action: func(c *cli.Context) error {
				logger.Debugf("Verify (cmd)")
				signerCheck := c.String("signer")

				if c.String("in") != "" {
					verified, _, err := verifyFileForCLI(c, client)
					if err != nil {
						return err
					}
					if signerCheck != "" {
						if err := checkSigner(verified, signerCheck); err != nil {
							return err
						}
					} else if verified != nil {
						fmtKey(os.Stdout, verified, "verified ")
					}
					return nil
				}

				reader := bufio.NewReader(os.Stdin)
				writer := os.Stdout

				sigFile := c.String("sig")
				if sigFile != "" {
					verified, err := verifyDetachedStream(client, c.Bool("armor"), reader, sigFile)
					if err != nil {
						return err
					}
					if signerCheck != "" {
						if err := checkSigner(verified, signerCheck); err != nil {
							return err
						}
					} else if verified != nil {
						fmtKey(os.Stdout, verified, "verified ")
					}
					return nil
				}

				logger.Debugf("Verify stream (cmd) armored=%t", c.Bool("armor"))
				verifyClient, err := NewVerifyStreamClient(context.TODO(), client.ProtoClient(), c.Bool("armor"))
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
						if signerCheck != "" {
							if err := checkSigner(resp.Signer, signerCheck); err != nil {
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

func verifyFileForCLI(c *cli.Context, client *Client) (*Key, string, error) {
	sigFile := c.String("sig")
	if sigFile != "" {
		signer, err := verifyDetachedFile(client, c.Bool("armor"), c.String("in"), sigFile)
		return signer, "", err
	}
	return verifyFile(client, c.Bool("armor"), c.String("in"), c.String("out"))
}

func verifyFile(client *Client, armored bool, in string, out string) (*Key, string, error) {
	logger.Debugf("Verify file (cmd) in=%s, out=%s, armored=%t", in, out, armored)
	if in == "" {
		return nil, "", errors.Errorf("in not specified")
	}
	in, err := filepath.Abs(in)
	if err != nil {
		return nil, "", err
	}
	if out != "" {
		out, err = filepath.Abs(out)
		if err != nil {
			return nil, "", err
		}
	}

	verifyClient, err := client.ProtoClient().VerifyFile(context.TODO())
	if err != nil {
		return nil, "", err
	}

	if err := verifyClient.Send(&VerifyFileInput{
		Armored: armored,
		In:      in,
		Out:     out,
	}); err != nil {
		return nil, "", err
	}

	resp, recvErr := verifyClient.CloseAndRecv()
	if recvErr != nil {
		return nil, "", recvErr
	}

	return resp.Signer, resp.Out, nil
}

func verifyDetachedFile(client *Client, armored bool, in string, sigFile string) (*Key, error) {
	logger.Debugf("Verify detached file (cmd) in=%s, sig=%s, armored=%t", in, sigFile, armored)
	if in == "" {
		return nil, errors.Errorf("in not specified")
	}
	in, err := filepath.Abs(in)
	if err != nil {
		return nil, err
	}

	sig, err := ioutil.ReadFile(sigFile)
	if err != nil {
		return nil, err
	}

	verifyClient, err := client.ProtoClient().VerifyDetachedFile(context.TODO())
	if err != nil {
		return nil, err
	}

	if err := verifyClient.Send(&VerifyDetachedFileInput{
		Armored: armored,
		In:      in,
		Sig:     sig,
	}); err != nil {
		return nil, err
	}

	resp, recvErr := verifyClient.CloseAndRecv()
	if recvErr != nil {
		return nil, recvErr
	}

	return resp.Signer, nil
}

func verifyDetachedStream(client *Client, armored bool, reader io.Reader, sigFile string) (*Key, error) {
	logger.Debugf("Verify detached stream (cmd) sig=%s, armored=%t", sigFile, armored)
	sig, err := ioutil.ReadFile(sigFile)
	if err != nil {
		return nil, err
	}

	verifyClient, err := client.ProtoClient().VerifyDetachedStream(context.TODO())
	if err != nil {
		return nil, err
	}

	sentSig := false
	var resp *VerifyDetachedResponse

	_, inErr := readFrom(reader, 1024*1024, func(b []byte) error {
		if len(b) > 0 {
			var req *VerifyDetachedInput
			if !sentSig {
				req = &VerifyDetachedInput{
					Armored: armored,
					Sig:     sig,
					Data:    b,
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
		return nil, inErr
	}

	if resp == nil {
		return nil, errors.Errorf("no response")
	}

	return resp.Signer, nil
}
