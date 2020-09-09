package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
	"unicode/utf8"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"google.golang.org/protobuf/proto"
)

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
		fmtKey(w, key)
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

func fmtKey(w io.Writer, key *Key) {
	if key == nil {
		return
	}
	fmt.Fprintf(w, key.ID)
	if key.User != nil {
		fmt.Fprint(w, " ")
		fmt.Fprint(w, fmtUser(key.User))
	}
	fmt.Fprintf(w, "\n")
}

func fmtVerifiedEncrypt(w io.Writer, key *Key, mode EncryptMode) {
	if key == nil {
		return
	}
	fmt.Fprint(w, "verified: ")
	fmt.Fprintf(w, key.ID)
	fmt.Fprintf(w, " %s", encryptModeToString(mode))
	if key.User != nil {
		fmt.Fprint(w, " ")
		fmt.Fprint(w, fmtUser(key.User))
	}
	fmt.Fprintf(w, "\n")
}

func fmtVerified(w io.Writer, key *Key) {
	if key == nil {
		return
	}
	fmt.Fprint(w, "verified: ")
	fmt.Fprintf(w, key.ID)
	if key.User != nil {
		fmt.Fprint(w, " ")
		fmt.Fprint(w, fmtUser(key.User))
	}
	fmt.Fprintf(w, "\n")
}

func encryptModeToString(m EncryptMode) string {
	switch m {
	case SaltpackEncrypt:
		return "saltpack-encrypt"
	case SaltpackSigncrypt:
		return "saltpack-signcrypt"
	default:
		return "unknown"
	}
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
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
