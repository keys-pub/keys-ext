package git

import (
	git "github.com/keys-pub/git2go"
	"github.com/pkg/errors"
)

// ErrorCode alias.
type ErrorCode = git.ErrorCode

// ErrNonFastForward alias.
const ErrNonFastForward ErrorCode = git.ErrNonFastForward

// ErrUnbornBranch alias.
const ErrUnbornBranch ErrorCode = git.ErrUnbornBranch

// ErrIsCode check if error is same as specified code.
func ErrIsCode(err error, code ErrorCode) bool {
	switch err := errors.Cause(err).(type) {
	case *git.GitError:
		return err.Code == code
	default:
		return false
	}
}
