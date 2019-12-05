package service

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func sigchainCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "sigchain",
			Usage: "Manage a sigchain",
			Subcommands: []cli.Command{
				cli.Command{
					Name:  "show",
					Usage: "Show sigchain",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "kid, k", Usage: "kid"},
						cli.IntFlag{Name: "seq, s", Usage: "seq"},
					},
					Action: func(c *cli.Context) error {
						seq := c.Int("seq")
						resp, err := client.ProtoClient().Sigchain(context.TODO(), &SigchainRequest{
							KID: c.String("kid"),
							Seq: int32(seq),
						})
						if err != nil {
							return err
						}

						if seq != 0 {
							if len(resp.Statements) == 0 {
								return errors.Errorf("no statement")
							}
							st, err := statementFromRPC(resp.Statements[0])
							if err != nil {
								return err
							}
							fmt.Println(string(st.Bytes()))
							return nil
						}

						logger.Infof("Resolving statements")
						sts, stsErr := statementsFromRPC(resp.Statements)
						if stsErr != nil {
							return errors.Wrapf(stsErr, "failed to resolve statements")
						}
						kid, err := keys.ParseID(resp.KID)
						if err != nil {
							return err
						}
						logger.Infof("Resolving sigchain from statements")
						sc, err := keys.NewSigchainForKID(kid)
						if err != nil {
							return err
						}
						if err := sc.AddAll(sts); err != nil {
							return errors.Wrapf(err, "failed to resolve sigchain from statements")
						}
						for _, st := range sts {
							fmt.Println(string(st.Bytes()))
						}
						// spew, err := keys.Spew(sc.EntryIterator(), nil)
						// if err != nil {
						// 	return err
						// }
						// fmt.Println(spew.String())
						return nil
					},
				},
				cli.Command{
					Name:  "statement",
					Usage: "Manage sigchain statements",
					Subcommands: []cli.Command{
						cli.Command{
							Name:      "add",
							Usage:     "Add a signed statement to a sigchain (from stdin)",
							ArgsUsage: "<stdin>",
							Flags: []cli.Flag{
								cli.StringFlag{Name: "kid, k"},
								cli.BoolFlag{Name: "dry-run"},
								cli.BoolFlag{Name: "local"},
							},
							Action: func(c *cli.Context) error {
								if c.NArg() > 0 {
									return errors.Errorf("input is from stdin, not as an argument")
								}

								r := bufio.NewReader(os.Stdin)
								b, err := ioutil.ReadAll(r)
								if err != nil {
									return err
								}
								if len(b) > 16*1024 {
									return errors.Errorf("sigchain data restricted to 16KB")
								}

								resp, err := client.ProtoClient().SigchainStatementCreate(context.TODO(), &SigchainStatementCreateRequest{
									KID:    c.String("kid"),
									DryRun: c.Bool("dry-run"),
									Local:  c.Bool("local"),
									Data:   b,
								})
								if err != nil {
									return err
								}
								st, err := statementFromRPC(resp.Statement)
								if err != nil {
									return err
								}
								fmt.Printf("%s\n", string(st.Bytes()))
								return nil
							},
						},
						cli.Command{
							Name:  "revoke",
							Usage: "Revoke a signed statement in a sigchain (from stdin)",
							Flags: []cli.Flag{
								cli.StringFlag{Name: "kid, k"},
								cli.IntFlag{Name: "seq, s"},
								cli.BoolFlag{Name: "local"},
							},
							Action: func(c *cli.Context) error {
								resp, err := client.ProtoClient().SigchainStatementRevoke(context.TODO(), &SigchainStatementRevokeRequest{
									KID:    c.String("kid"),
									Seq:    int32(c.Int("seq")),
									DryRun: c.Bool("dry-run"),
									Local:  c.Bool("local"),
								})
								if err != nil {
									return err
								}
								st, err := statementFromRPC(resp.Statement)
								if err != nil {
									return err
								}
								fmt.Printf("%s\n", string(st.Bytes()))
								return nil
							},
						},
					},
				},
			},
		},
	}
}
