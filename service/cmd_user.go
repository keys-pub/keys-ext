package service

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/urfave/cli"
)

func userCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "user",
			Usage: "Manage users",
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
						resp, err := client.KeysClient().User(context.TODO(), &UserRequest{
							KID:   kid,
							Local: c.Bool("local"),
						})
						if err != nil {
							return err
						}
						if resp.User != nil {
							fmt.Println(fmtUser(resp.User))
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
							fmt.Fprintf(w, "%s\t%s\n", fmtUser(user), user.KID)
						}
						if err := w.Flush(); err != nil {
							return err
						}
						fmt.Print(out.String())
						return nil
					},
				},
				cli.Command{
					Name:  "setup",
					Usage: "Link a key to an account on Twitter or Github",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "kid, k", Usage: "key (defaults to current key)"},
					},
					Action: func(c *cli.Context) error {
						kid, err := argString(c, "kid", false)
						if err != nil {
							return err
						}

						// TODO: Reddit, etc
						reader := bufio.NewReader(os.Stdin)
						fmt.Println("What's the service? ")
						fmt.Println("(g) Github")
						fmt.Println("(t) Twitter")
						sservice, err := reader.ReadString('\n')
						if err != nil {
							return err
						}

						service := ""
						question := ""
						switch strings.TrimSpace(strings.ToLower(sservice)) {
						case "g", "github":
							service = "github"
							question = "What's your Github username?"
							// next = "In the next step, we'll create a signed message that you can post as a gist on your Github account."
						case "t", "twitter":
							service = "twitter"
							question = "What's your Twitter handle?"
							// next = "In the next step, we'll create a signed message that you can post as a tweet."
						}

						fmt.Println("")
						fmt.Print(question + " ")
						uin, err := reader.ReadString('\n')
						if err != nil {
							return err
						}
						name := strings.TrimSpace(strings.ToLower(uin))

						signResp, err := client.KeysClient().UserSign(context.TODO(), &UserSignRequest{
							KID:     kid,
							Service: service,
							Name:    name,
						})
						if err != nil {
							return err
						}
						instructions := ""
						link := ""
						urlq := ""
						switch service {
						case "github":
							instructions = "Create a new gist on your Github account, and paste the signed message there."
							link = "https://gist.github.com/new"
							urlq = "What's the location (URL) on github.com where the signed message was saved?"
						case "twitter":
							instructions = "Save the following signed message as a tweet on your Twitter account."
							link = "https://twitter.com/intent/tweet"
							urlq = "What's the location (URL) on twitter.com where the signed message tweet was saved?"
						}
						fmt.Println("")
						fmt.Println(instructions)
						fmt.Println(link)
						fmt.Println("")
						fmt.Println(signResp.Message)
						fmt.Println("")

						fmt.Print("Have you posted the signed message (Y/n)? ")
						proceed, err := reader.ReadString('\n')
						if err != nil {
							return err
						}
						switch strings.TrimSpace(strings.ToLower(proceed)) {
						case "y", "yes", "":
						default:
							return nil
						}

						fmt.Print(urlq + " ")
						surl, err := reader.ReadString('\n')
						if err != nil {
							return err
						}
						url := strings.TrimSpace(strings.ToLower(surl))

						_, err = client.KeysClient().UserAdd(context.TODO(), &UserAddRequest{
							KID:     kid,
							Service: service,
							Name:    name,
							URL:     url,
						})
						if err != nil {
							return err
						}
						// fmt.Printf("%s %s %s\n", setResp.User.KID, setResp.User.DisplayName, setResp.User.URL)
						fmt.Println("User successfully setup.")
						return nil
					},
				},
				cli.Command{
					Name:      "sign",
					Usage:     "Create a signed user statement",
					ArgsUsage: "",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "kid", Usage: "kid (defaults to current key)"},
						cli.StringFlag{Name: "service"},
						cli.StringFlag{Name: "name"},
					},
					Action: func(c *cli.Context) error {
						resp, err := client.KeysClient().UserSign(context.TODO(), &UserSignRequest{
							KID:     c.String("kid"),
							Service: c.String("service"),
							Name:    c.String("name"),
						})
						if err != nil {
							return err
						}
						fmt.Println(resp.Message)
						return nil
					},
				},
				cli.Command{
					Name:      "add",
					Usage:     "Add a verified user statement to the sigchain",
					ArgsUsage: "",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "kid", Usage: "kid (defaults to current key)"},
						cli.StringFlag{Name: "service"},
						cli.StringFlag{Name: "name"},
						cli.StringFlag{Name: "url", Usage: "URL to signed statement created by `keys user sign`"},
						cli.BoolFlag{Name: "local", Usage: "Don't save to the key server"},
					},
					Action: func(c *cli.Context) error {
						resp, err := client.KeysClient().UserAdd(context.TODO(), &UserAddRequest{
							KID:     c.String("kid"),
							Service: c.String("service"),
							Name:    c.String("name"),
							URL:     c.String("url"),
							Local:   c.Bool("local"),
						})
						if err != nil {
							return err
						}
						st, err := statementFromRPC(resp.Statement)
						if err != nil {
							return err
						}
						b, err := st.Bytes()
						if err != nil {
							return err
						}
						fmt.Printf("%s\n", string(b))
						return nil
					},
				},
			},
		},
	}
}
