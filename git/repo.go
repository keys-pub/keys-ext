package git

import (
	"fmt"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"

	"github.com/pkg/errors"
)

// Repository ...
type Repository struct {
	repo *git.Repository
	path string
	opts Options

	cache *files
}

// NewRepository creates a new repository.
func NewRepository(opt ...Option) (*Repository, error) {
	opts, err := newOptions(opt...)
	if err != nil {
		return nil, err
	}
	return &Repository{opts: opts}, nil
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

	remote, err := repo.Remote("origin")
	if err != nil {
		return err
	}
	logger.Infof("Remote: %s", remote.Config().URLs)

	r.repo = repo
	r.path = path
	return nil
}

// Init to create an empty git repo.
func (r *Repository) Init(path string) error {
	repo, err := git.PlainInit(path, false)
	if err != nil {
		return err
	}
	r.repo = repo
	r.path = path
	return nil
}

// Clone repository.
func (r *Repository) Clone(urs string, path string) error {
	if r.repo != nil {
		return errors.Errorf("already open")
	}

	// Prepare temp path
	tmpPath := path + ".clone"
	tmpExists, err := pathExists(tmpPath)
	if err != nil {
		return err
	}
	if tmpExists {
		if err := os.RemoveAll(tmpPath); err != nil {
			return err
		}
	}
	defer func() { _ = os.RemoveAll(tmpPath) }()

	// Clone to temp
	empty := false
	logger.Debugf("Clone repo: %s", urs)
	if _, err := git.PlainClone(tmpPath, false, &git.CloneOptions{
		URL: urs,
		// Progress: os.Stdout,
		Auth: r.opts.auth,
	}); err != nil {
		if errors.Cause(err) == transport.ErrEmptyRemoteRepository {
			logger.Infof("Repository is empty")
			empty = true
		} else {
			return err
		}
	}

	if empty {
		logger.Infof("Initializing repo: %s", tmpPath)
		repo, err := git.PlainInit(tmpPath, false)
		if err != nil {
			return err
		}
		cfg := &config.RemoteConfig{
			Name: "origin",
			URLs: []string{urs},
		}
		if _, err := repo.CreateRemote(cfg); err != nil {
			return err
		}
	}

	// Move into place
	logger.Debugf("Moving repo into place: %s", path)
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}

	return r.Open(path)
}

func (r *Repository) fetch() error {
	if r.repo == nil {
		return errors.Errorf("not open")
	}
	logger.Debugf("Fetch origin...")
	fetchOptions := &git.FetchOptions{
		RemoteName: "origin",
		Auth:       r.opts.auth,
	}
	if err := r.repo.Fetch(fetchOptions); err != nil {
		if errors.Cause(err) == git.NoErrAlreadyUpToDate {
			logger.Debugf("Fetch (already up to date)")
			return nil
		}
		return err
	}

	return nil
}

// Pull fetches and merges.
func (r *Repository) Pull() error {
	return r.pull()
}

func (r *Repository) pull() error {
	logger.Debugf("Pull...")
	tree, err := r.repo.Worktree()
	if err != nil {
		return err
	}

	if err := tree.Pull(&git.PullOptions{
		Auth: r.opts.auth,
	}); err != nil {
		if errors.Cause(err) == git.NoErrAlreadyUpToDate {
			logger.Debugf("Push (already up to date)")
			return nil
		}
		return err
	}

	logger.Debugf("Pull complete")
	return nil

}

// Push changes.
func (r *Repository) Push() error {
	return r.push()
}

func (r *Repository) push() error {
	if r.repo == nil {
		return errors.Errorf("not open")
	}

	logger.Debugf("Push git")
	if err := r.repo.Push(&git.PushOptions{
		Auth: r.opts.auth,
	}); err != nil {
		if errors.Cause(err) == git.NoErrAlreadyUpToDate {
			logger.Debugf("Push (already up to date)")
			return nil
		}
		return err
	}
	logger.Debugf("Push complete")
	return nil
}

// Sync does a pull (fetch, merge), push.
func (r *Repository) Sync() error {
	logger.Infof("Syncing git remote...")

	if err := r.pull(); err != nil {
		return errors.Wrapf(err, "failed to sync (pull)")
	}
	if err := r.push(); err != nil {
		return errors.Wrapf(err, "failed to sync (push)")
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

func (r *Repository) addCommit(name string, msg string) error {
	logger.Debugf("Git add: %s", msg)

	w, err := r.repo.Worktree()
	if err != nil {
		return err
	}

	if _, err := w.Add(name); err != nil {
		return err
	}

	commit, err := w.Commit(msg, &git.CommitOptions{
		Author: r.signature(),
	})
	logger.Debugf("Commit %s", commit)
	return nil
}

func (r *Repository) removeCommit(name string) error {
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
