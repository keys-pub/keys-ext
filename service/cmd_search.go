package service

import (
	"bytes"
	"context"

	"fmt"
	"text/tabwriter"

	"github.com/urfave/cli"
)

func searchCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "search",
			Usage: "Search",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "query, q", Usage: "query"},
				cli.IntFlag{Name: "limit, l", Usage: "limit number of results"},
			},
			Action: func(c *cli.Context) error {
				query, err := argString(c, "query", true)
				if err != nil {
					return err
				}
				searchResp, err := client.ProtoClient().Search(context.TODO(), &SearchRequest{
					Query: query,
					Limit: int32(c.Int("limit")),
				})
				if err != nil {
					return err
				}

				out := &bytes.Buffer{}
				w := new(tabwriter.Writer)
				w.Init(out, 0, 8, 1, ' ', 0)
				for _, res := range searchResp.Results {
					fmtResult(w, res)
				}
				w.Flush()
				fmt.Printf(out.String())
				return nil
			},
		},
	}
}
