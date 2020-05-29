package git

import (
	"fmt"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// Repository ...
type Repository struct {
	repo *git.Repository
	path string
	auth transport.AuthMethod
}

// NewRepository ...
func NewRepository() *Repository {
	return &Repository{}
}

// Path to repo.
func (r *Repository) Path() string {
	return r.path
}

func pathExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

// Open repository.
func (r *Repository) Open(path string) error {
	if r.repo != nil {
		return errors.Errorf("already open")
	}

	logger.Debugf("Open repo: %s", path)
	exist, err := pathExists(path)
	if err != nil {
		return err
	}
	if !exist {
		return errors.Errorf("path doesn't exist: %s", r.path)
	}

	repo, err := git.PlainOpen(path)
	if err != nil {
		return err
	}
	r.repo = repo
	r.path = path
	logger.Debugf("Opened repo: %s", path)
	return nil
}

// SetKey sets the ssh key.
func (r *Repository) SetKey(key *keys.EdX25519Key) error {
	// privateKey, err := key.EncodeToSSH(nil)
	// if err != nil {
	// 	return err
	// }
	// signer, err := ssh.ParsePrivateKey(privateKey)
	// if err != nil {
	// 	return err

	r.auth = &gitssh.PublicKeys{User: "git", Signer: key.SSHSigner()}
	return nil
}

// Clone repository.
func (r *Repository) Clone(urs string, path string) error {
	if r.repo != nil {
		return errors.Errorf("already open")
	}

	// Prepare temp path
	tmpPath := path + ".tmp"
	tmpExists, err := pathExists(tmpPath)
	if err != nil {
		return err
	}
	if tmpExists {
		if err := os.RemoveAll(tmpPath); err != nil {
			return err
		}
	}

	// Clone to temp
	logger.Debugf("Clone repo: %s", urs)
	if _, err := git.PlainClone(tmpPath, false, &git.CloneOptions{
		URL: urs,
		// Progress: os.Stdout,
		Auth: r.auth,
	}); err != nil {
		return err
	}

	// Move into place
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}

	// Open repo
	repo, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	r.repo = repo
	r.path = path
	logger.Debugf("Cloned repo: %s", r.path)
	return nil
}

// // Pull changes.
// func (r *Repository) Pull() error {
// 	if r.repo == nil {
// 		return errors.Errorf("not open")
// 	}

// 	w, err := r.repo.Worktree()
// 	if err != nil {
// 		return err
// 	}

// 	if err := w.Pull(&git.PullOptions{
// 		RemoteName: "origin",
// 		Auth:       r.auth,
// 	}); err != nil {
// 		return err
// 	}

// 	return nil
// }

// Fetch remote.
func (r *Repository) Fetch() error {
	if r.repo == nil {
		return errors.Errorf("not open")
	}

	logger.Debugf("Fetch origin...")
	fetchOptions := &git.FetchOptions{
		RemoteName: "origin",
		Auth:       r.auth,
	}
	if err := r.repo.Fetch(fetchOptions); err != nil {
		return err
	}

	return nil
}

// Pull fetches and merges.
func (r *Repository) Pull() error {
	if err := r.Fetch(); err != nil {
		return err
	}
	if err := r.Merge(); err != nil {
		return err
	}
	return nil
}

// Push changes.
func (r *Repository) Push() error {
	if r.repo == nil {
		return errors.Errorf("not open")
	}

	if err := r.repo.Push(&git.PushOptions{
		Auth: r.auth,
	}); err != nil {
		return err
	}
	return nil
}

func (r *Repository) signature() *object.Signature {
	return &object.Signature{
		Name:  "keys.pub",
		Email: "git@keys.pub",
		When:  time.Now(),
	}
}

// Add file.
func (r *Repository) Add(name string) error {
	logger.Debugf("Add %s", name)
	message := fmt.Sprintf("Add %s\n", name)

	w, err := r.repo.Worktree()
	if err != nil {
		return err
	}

	if _, err := w.Add(name); err != nil {
		return err
	}

	commit, err := w.Commit(message, &git.CommitOptions{
		Author: r.signature(),
	})
	logger.Debugf("Commit %s", commit)
	return nil
}

// Remove file.
func (r *Repository) Remove(name string) error {
	logger.Debugf("Remove %s", name)
	message := fmt.Sprintf("Remove %s\n", name)

	w, err := r.repo.Worktree()
	if err != nil {
		return err
	}

	if _, err := w.Remove(name); err != nil {
		return err
	}

	commit, err := w.Commit(message, &git.CommitOptions{
		Author: r.signature(),
	})
	logger.Debugf("Commit %s", commit)
	return nil
}
