package service

import (
	"bytes"
	"context"

	"fmt"
	"text/tabwriter"

	"github.com/pkg/errors"
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
				if c.NArg() > 0 {
					return errors.Errorf("too many arguments")
				}
				searchResp, err := client.ProtoClient().Search(context.TODO(), &SearchRequest{
					Query: c.String("query"),
					Limit: int32(c.Int("limit")),
				})
				if err != nil {
					return err
				}

				out := &bytes.Buffer{}
				w := new(tabwriter.Writer)
				w.Init(out, 0, 8, 1, ' ', 0)
				for _, key := range searchResp.Keys {
					fmtKey(w, key)
				}
				w.Flush()
				fmt.Printf(out.String())
				return nil
			},
		},
	}
}

// func fmtSearchResult(w io.Writer, res *SearchResult) {
// 	if res == nil {
// 		return
// 	}
// 	typ := ""
// 	if res.Type == PrivateKeyType {
// 		typ = "🔑"
// 	}
// 	fmt.Fprintf(w, "%s\t%s\t%s\n", res.KID, fmtUsers(res.Users), typ)
// }
