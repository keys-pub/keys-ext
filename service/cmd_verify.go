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
				cli.StringFlag{Name: "mode, m", Value: "", Usage: "mode: armor, binary, attached, detached"},
				cli.StringFlag{Name: "sig, x", Usage: "signature file (if detached)"},
				cli.StringFlag{Name: "in, i", Usage: "file to read"},
				cli.StringFlag{Name: "out, o", Usage: "file to write (if attached), defaults to {in} without .signed"},
			},
			Action: func(c *cli.Context) error {
				logger.Debugf("Verify (cmd)")

				// TODO: Error if out path already exists

				mode, err := parseMode(c.String("mode"), c.String("sig") != "")
				if err != nil {
					return err
				}

				signer := c.String("signer")
				if signer == "" {
					return errors.Errorf("specify -s (-signer) to verify")
				}

				if c.String("in") != "" {
					if _, err := verifyFileForCLI(c, client, mode, signer); err != nil {
						return err
					}
					return nil
				}

				reader := bufio.NewReader(os.Stdin)
				writer := os.Stdout

				sigFile := c.String("sig")
				if sigFile != "" {
					if !mode.isDetached(stdIn) {
						return errors.Errorf("sig is only for detached mode")
					}
					if err := verifyDetachedStream(client, mode.isArmored(stdIn, false), reader, sigFile, signer); err != nil {
						return err
					}
					return nil
				}

				if mode.isDetached(stdIn) {
					return errors.Errorf("detached mode without sig")
				}

				armored := mode.isArmored(stdIn, false)
				logger.Debugf("Verify stream (cmd) armored=%t, detatched=false", armored)
				verifyClient, err := NewVerifyStreamClient(context.TODO(), client.ProtoClient(), armored)
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
						if err := checkSigner(resp.Signer, signer); err != nil {
							outErr = err
							break
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

func verifyFileForCLI(c *cli.Context, client *Client, mode signMode, signer string) (string, error) {
	sigFile := c.String("sig")
	in := c.String("in")
	detached := mode.isDetached(fileIn)

	if detached {
		if sigFile == "" {
			sigFile = in + ".sig"
		}
		if err := verifyDetachedFile(client, mode.isArmored(fileIn, detached), in, sigFile, signer); err != nil {
			return "", err
		}
		return "", nil
	}

	return verifyFile(client, mode.isArmored(fileIn, detached), in, c.String("out"), signer)
}

func verifyFile(client *Client, armored bool, in string, out string, signer string) (string, error) {
	logger.Debugf("Verify file (cmd) in=%s, out=%s, armored=%t, detached=false", in, out, armored)
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

	verifyClient, err := client.ProtoClient().VerifyFile(context.TODO())
	if err != nil {
		return "", err
	}

	if err := verifyClient.Send(&VerifyFileInput{
		Armored: armored,
		In:      in,
		Out:     out,
	}); err != nil {
		return "", err
	}

	resp, recvErr := verifyClient.CloseAndRecv()
	if recvErr != nil {
		return "", recvErr
	}

	if err := checkSigner(resp.Signer, signer); err != nil {
		return "", err
	}

	return resp.Out, nil
}

func verifyDetachedFile(client *Client, armored bool, in string, sigFile string, signer string) error {
	logger.Debugf("Verify detached file (cmd) in=%s, sig=%s, armored=%t", in, sigFile, armored)
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

	verifyClient, err := client.ProtoClient().VerifyDetachedFile(context.TODO())
	if err != nil {
		return err
	}

	if err := verifyClient.Send(&VerifyDetachedFileInput{
		Armored: armored,
		In:      in,
		Sig:     sig,
	}); err != nil {
		return err
	}

	resp, recvErr := verifyClient.CloseAndRecv()
	if recvErr != nil {
		return recvErr
	}

	if err := checkSigner(resp.Signer, signer); err != nil {
		return err
	}

	return nil
}

func verifyDetachedStream(client *Client, armored bool, reader io.Reader, sigFile string, signer string) error {
	logger.Debugf("Verify detached stream (cmd) sig=%s, armored=%t", sigFile, armored)
	sig, err := ioutil.ReadFile(sigFile) // #nosec
	if err != nil {
		return err
	}

	verifyClient, err := client.ProtoClient().VerifyDetachedStream(context.TODO())
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
		return inErr
	}

	if resp == nil {
		return errors.Errorf("no response")
	}

	// TODO: Send expected signer in request
	if err := checkSigner(resp.Signer, signer); err != nil {
		return err
	}

	return nil
}

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
