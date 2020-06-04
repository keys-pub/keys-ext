package service

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"unicode/utf8"

	proto "github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/jsonpb"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func readerFromArgs(in string) (io.Reader, error) {
	if in != "" {
		file, err := os.Open(in) // #nosec
		if err != nil {
			return nil, err
		}
		return bufio.NewReader(file), nil
	}
	return bufio.NewReader(os.Stdin), nil
}

func writerFromArgs(out string) (io.Writer, error) {
	if out != "" {
		file, createErr := os.Create(out)
		if createErr != nil {
			return nil, createErr
		}
		return file, nil
	}
	return os.Stdout, nil
}

func writeAll(writer io.Writer, b []byte) error {
	n, writeErr := writer.Write(b)
	if writeErr != nil {
		return writeErr
	}
	if n != len(b) {
		return io.ErrShortWrite
	}
	return nil
}

func fmtKeys(keys []*Key) {
	out := &bytes.Buffer{}
	w := new(tabwriter.Writer)
	w.Init(out, 0, 8, 1, ' ', 0)
	for _, key := range keys {
		fmtKey(w, key, "")
	}
	if err := w.Flush(); err != nil {
		panic(err)
	}
	fmt.Print(out.String())
}

func fmtUser(user *User) string {
	if user == nil {
		return ""
	}
	s := fmt.Sprintf("%s@%s", user.Name, user.Service)
	switch user.Status {
	case UserStatusOK:
		return s
	case UserStatusUnknown:
		return fmt.Sprintf("unknown:%s", s)
	case UserStatusConnFailure:
		return fmt.Sprintf("connfail:%s", s)
	default:
		return fmt.Sprintf("failed:%s", s)
	}
}

func fmtKey(w io.Writer, key *Key, prefix string) {
	if key == nil {
		fmt.Fprintf(w, "∅\n")
		return
	}
	if prefix != "" {
		fmt.Fprint(w, prefix)
	}
	fmt.Fprintf(w, key.ID)
	if key.User != nil {
		fmt.Fprint(w, " ")
		fmt.Fprint(w, fmtUser(key.User))
	}
	fmt.Fprintf(w, "\n")
}

func fmtItems(items []*Item) {
	out := &bytes.Buffer{}
	w := new(tabwriter.Writer)
	w.Init(out, 0, 8, 1, ' ', 0)
	for _, item := range items {
		fmtItem(w, item)
	}
	if err := w.Flush(); err != nil {
		panic(err)
	}
	fmt.Print(out.String())
}

func fmtItem(w io.Writer, item *Item) {
	if item == nil {
		return
	}
	fmt.Fprintf(w, "%s\t%s\n", item.ID, item.Type)
}

func fmtContent(w io.Writer, content *Content) {
	switch content.Type {
	case UTF8Content:
		if utf8.Valid(content.Data) {
			fmt.Fprintf(w, "%s", string(content.Data))
		} else {
			fmt.Fprintf(w, "[invalid utf8]")
		}
	default:
		fmt.Fprintf(w, "[bytes len(%d)]", len(content.Data))
	}
}

func identityForKey(k *Key) string {
	if k.User != nil {
		return k.User.ID
	}
	return k.ID
}

func fmtMessage(w io.Writer, msg *Message) {
	if msg == nil || msg.Content == nil || len(msg.Content.Data) == 0 {
		return
	}
	fmt.Fprintf(w, "%s: ", identityForKey(msg.Sender))
	fmtContent(w, msg.Content)
	fmt.Fprintf(w, "\n")
}

func argString(c *cli.Context, name string, optional bool) (string, error) {
	val := c.String(name)
	if val != "" {
		if c.NArg() > 0 {
			return "", errors.Errorf("too many arguments")
		}
		return val, nil
	}

	args := c.Args().First()
	if args != "" {
		return args, nil
	}

	if optional {
		return "", nil
	}

	return "", errors.Errorf("no %s specified", name)
}

func printMessage(m proto.Message) {
	marshal := jsonpb.Marshaler{
		Indent: "  ",
	}
	b, err := marshal.MarshalToString(m)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
