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
		{
			Name:  "user",
			Usage: "Link and search users",
			Subcommands: []cli.Command{
				{
					Name:  "find",
					Usage: "Find by kid",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "kid, k", Usage: "key"},
						cli.BoolFlag{Name: "local", Usage: "search local index only"},
					},
					Action: func(c *cli.Context) error {
						kid, err := argString(c, "kid", false)
						if err != nil {
							return err
						}
						resp, err := client.RPCClient().User(context.TODO(), &UserRequest{
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
				{
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
						searchResp, err := client.RPCClient().UserSearch(context.TODO(), &UserSearchRequest{
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
				{
					Name:  "setup",
					Usage: "Link a key to an account (Twitter, Github, Reddit)",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "kid, k", Usage: "key"},
					},
					Action: func(c *cli.Context) error {
						kid, err := argString(c, "kid", false)
						if err != nil {
							return err
						}

						reader := bufio.NewReader(os.Stdin)
						fmt.Println("What's the service? ")
						fmt.Println("(g) Github")
						fmt.Println("(t) Twitter")
						fmt.Println("(r) Reddit")
						input, err := reader.ReadString('\n')
						if err != nil {
							return err
						}

						service := ""
						question := ""
						switch strings.TrimSpace(strings.ToLower(input)) {
						case "g", "github":
							service = "github"
							question = "What's your Github username?"
						case "t", "twitter":
							service = "twitter"
							question = "What's your Twitter handle?"
						case "r", "reddit":
							service = "reddit"
							question = "What's your Reddit username?"
						}

						fmt.Println("")
						fmt.Print(question + " ")
						uin, err := reader.ReadString('\n')
						if err != nil {
							return err
						}
						name := strings.TrimSpace(strings.ToLower(uin))

						signResp, err := client.RPCClient().UserSign(context.TODO(), &UserSignRequest{
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
						case "reddit":
							instructions = "Save the following signed message as a post on /r/keyspubmsgs."
							link = "https://old.reddit.com/r/keyspubmsgs/submit"
							urlq = "What's the location (URL) on reddit.com/r/keyspubmsgs where the signed message was posted?"
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

						_, err = client.RPCClient().UserAdd(context.TODO(), &UserAddRequest{
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
				{
					Name:      "sign",
					Usage:     "Create a signed user statement",
					ArgsUsage: "",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "kid, k", Usage: "key"},
						cli.StringFlag{Name: "service"},
						cli.StringFlag{Name: "name"},
					},
					Action: func(c *cli.Context) error {
						resp, err := client.RPCClient().UserSign(context.TODO(), &UserSignRequest{
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
				{
					Name:      "add",
					Usage:     "Add a verified user statement to a sigchain",
					ArgsUsage: "",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "kid, k", Usage: "key"},
						cli.StringFlag{Name: "service"},
						cli.StringFlag{Name: "name"},
						cli.StringFlag{Name: "url", Usage: "URL to signed statement created by `keys user sign`"},
						cli.BoolFlag{Name: "local", Usage: "Don't save to the key server"},
					},
					Action: func(c *cli.Context) error {
						resp, err := client.RPCClient().UserAdd(context.TODO(), &UserAddRequest{
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
