package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func dbCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "db",
			Usage: "DB",
			Subcommands: []cli.Command{
				cli.Command{
					Name:  "collections",
					Usage: "List collections",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "db", Value: "service", Usage: "db"},
					},
					ArgsUsage: "",
					Hidden:    true,
					Action: func(c *cli.Context) error {
						if c.NArg() > 1 {
							return errors.Errorf("too many arguments")
						}
						path := c.Args().First()
						req := &CollectionsRequest{
							Parent: path,
							DB:     c.String("db"),
						}
						resp, err := client.KeysClient().Collections(context.TODO(), req)
						if err != nil {
							return err
						}
						fmtCollections(resp.Collections)
						return nil
					},
				},
				cli.Command{
					Name:  "documents",
					Usage: "List documents",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "db", Value: "service", Usage: "db"},
					},
					ArgsUsage: "collection",
					Hidden:    true,
					Subcommands: []cli.Command{
						documentDeleteCommand(client),
					},
					Action: func(c *cli.Context) error {
						if c.NArg() > 1 {
							return errors.Errorf("too many arguments")
						}
						path := strings.TrimSpace(c.Args().First())

						if path == "" {
							resp, err := client.KeysClient().Collections(context.TODO(), &CollectionsRequest{
								DB: c.String("db"),
							})
							if err != nil {
								return err
							}
							fmtCollections(resp.Collections)
							return nil
						}

						req := &DocumentsRequest{
							Prefix: path,
							DB:     c.String("db"),
						}
						resp, err := client.KeysClient().Documents(context.TODO(), req)
						if err != nil {
							return err
						}
						fmtDocuments(resp.Documents)
						return nil
					},
				},
			},
		},
	}
}

func documentDeleteCommand(client *Client) cli.Command {
	return cli.Command{
		Name:  "rm",
		Usage: "Delete document",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "db", Value: "service", Usage: "db"},
		},
		Hidden: true,
		Action: func(c *cli.Context) error {
			_, err := client.KeysClient().DocumentDelete(context.TODO(), &DocumentDeleteRequest{
				Path: c.Args().First(),
				DB:   c.String("db"),
			})
			if err != nil {
				return err
			}
			return nil
		},
	}
}

func fmtCollections(cols []*Collection) {
	out := &bytes.Buffer{}
	w := new(tabwriter.Writer)
	w.Init(out, 0, 8, 1, ' ', 0)
	for _, col := range cols {
		fmtCollection(w, col)
	}
	if err := w.Flush(); err != nil {
		panic(err)
	}
	fmt.Print(out.String())
}

func fmtCollection(w io.Writer, col *Collection) {
	if col == nil {
		fmt.Fprintf(w, "∅\n")
		return
	}
	fmt.Fprintf(w, "%s\n", col.Path)
}

func fmtDocuments(docs []*Document) {
	out := &bytes.Buffer{}
	w := new(tabwriter.Writer)
	w.Init(out, 0, 8, 1, ' ', 0)
	for _, doc := range docs {
		fmtDocument(w, doc)
	}
	if err := w.Flush(); err != nil {
		panic(err)
	}
	fmt.Print(out.String())
}

func fmtDocument(w io.Writer, doc *Document) {
	if doc == nil {
		fmt.Fprintf(w, "∅\n")
		return
	}
	fmt.Fprintf(w, "%s\t%s\n", doc.Path, doc.Value)
}
