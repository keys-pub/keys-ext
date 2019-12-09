package service

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func sigchainCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "sigchain",
			Usage: "Sigchains",
			Subcommands: []cli.Command{
				cli.Command{
					Name:  "show",
					Usage: "Show sigchain",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "kid, k", Usage: "kid"},
						cli.IntFlag{Name: "seq, s", Usage: "seq"},
					},
					Action: func(c *cli.Context) error {
						kid, err := argString(c, "kid", true)
						if err != nil {
							return err
						}
						seq := c.Int("seq")
						if seq != 0 {
							resp, err := client.ProtoClient().Statement(context.TODO(), &StatementRequest{
								KID: kid,
								Seq: int32(seq),
							})
							if err != nil {
								return err
							}
							st, err := statementFromRPC(resp.Statement)
							if err != nil {
								return err
							}
							fmt.Println(string(st.Bytes()))
							return nil
						}

						resp, err := client.ProtoClient().Sigchain(context.TODO(), &SigchainRequest{
							KID: kid,
						})
						if err != nil {
							return err
						}
						logger.Infof("Resolving statements")
						sc, err := sigchainFromRPC(resp.Key.KID, resp.Statements)
						if err != nil {
							return err
						}
						for _, st := range sc.Statements() {
							fmt.Println(string(st.Bytes()))
						}
						return nil
					},
				},
				cli.Command{
					Name:  "statement",
					Usage: "Sigchain statements",
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

								resp, err := client.ProtoClient().StatementCreate(context.TODO(), &StatementCreateRequest{
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
								resp, err := client.ProtoClient().StatementRevoke(context.TODO(), &StatementRevokeRequest{
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
