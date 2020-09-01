package service

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/urfave/cli"
)

func usersCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "users",
			Usage: "Find users",
			Subcommands: []cli.Command{
				cli.Command{
					Name:  "find",
					Usage: "Find by kid",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "kid, k", Usage: "kid"},
						cli.BoolFlag{Name: "local", Usage: "search local index only"},
					},
					Action: func(c *cli.Context) error {
						kid, err := argString(c, "kid", false)
						if err != nil {
							return err
						}
						resp, err := client.KeysClient().Users(context.TODO(), &UsersRequest{
							KID:   kid,
							Local: c.Bool("local"),
						})
						if err != nil {
							return err
						}
						if len(resp.Users) > 0 {
							fmtUsers(os.Stdout, resp.Users, "\n")
						}
						return nil
					},
				},
				cli.Command{
					Name:  "search",
					Usage: "Search for users",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "query, q", Usage: "query"},
						cli.IntFlag{Name: "limit, l", Usage: "limit number of results"},
						cli.BoolFlag{Name: "local", Usage: "search local index only"},
					},
					Action: func(c *cli.Context) error {
						query, err := argString(c, "query", true)
						if err != nil {
							return err
						}
						searchResp, err := client.KeysClient().UserSearch(context.TODO(), &UserSearchRequest{
							Query: query,
							Limit: int32(c.Int("limit")),
							Local: c.Bool("local"),
						})
						if err != nil {
							return err
						}

						out := &bytes.Buffer{}
						w := new(tabwriter.Writer)
						w.Init(out, 0, 8, 1, ' ', 0)
						for _, user := range searchResp.Users {
							fmtUser(w, user)
							fmt.Fprintf(w, "\t%s\n", user.KID)
						}
						if err := w.Flush(); err != nil {
							return err
						}
						fmt.Print(out.String())
						return nil
					},
				},
			},
		},
	}
}
