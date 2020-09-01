package service

import (
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/gogo/protobuf/jsonpb"
	proto "github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
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

func fmtKeys(w io.Writer, keys []*Key) {
	for _, key := range keys {
		fmtKey(w, key)
		fmt.Fprintf(w, "\n")
	}
}

func fmtUsers(w io.Writer, users []*User, delimeter string) {
	for i, user := range users {
		fmtUser(w, user)
		if i < len(users)-1 {
			fmt.Fprintf(w, delimeter)
		}
	}
}

func fmtUser(w io.Writer, user *User) {
	if user == nil {
		return
	}
	s := fmt.Sprintf("%s@%s", user.Name, user.Service)
	switch user.Status {
	case UserStatusOK:
		fmt.Fprintf(w, s)
	case UserStatusUnknown:
		fmt.Fprintf(w, "unknown:%s", s)
	case UserStatusConnFailure:
		fmt.Fprintf(w, "connfail:%s", s)
	default:
		fmt.Fprintf(w, "failed:%s", s)
	}
}

func fmtKey(w io.Writer, key *Key) {
	if key == nil {
		return
	}
	fmt.Fprintf(w, key.ID)
	if len(key.Users) > 0 {
		fmt.Fprint(w, " ")
	}
	fmtUsers(w, key.Users, ",")
}

func fmtVerified(w io.Writer, key *Key) {
	if key == nil {
		return
	}
	fmt.Fprint(w, "verified: ")
	fmtKey(w, key)
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
	if len(k.Users) > 0 {
		return k.Users[0].ID
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
