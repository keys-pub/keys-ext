package syncp

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// Git describes how to run git syncing.
type Git struct {
	repo string
}

// NewGit creates git syncing command.
func NewGit(repo string) (*Git, error) {
	return &Git{repo: repo}, nil
}

// Remote repository.
func (g *Git) Remote() string {
	return g.repo
}

func (g *Git) setup(cfg Config, rt Runtime) error {
	bin, err := exec.LookPath("git")
	if err != nil {
		return err
	}

	isEmpty := false
	res := Run(NewCmd(bin, Args("ls-remote", "-h", g.repo), Chdir(cfg.Dir)), rt)
	if res.Err != nil {
		return res.Err
	}
	refs := string(res.Output.Stdout)
	if refs == "" {
		isEmpty = true
	}

	if isEmpty {
		if err := ensureIgnore(cfg.Dir, rt); err != nil {
			return err
		}
		if res := Run(NewCmd(bin, Args("init"), Chdir(cfg.Dir)), rt); res.Err != nil {
			return res.Err
		}
		if res := Run(NewCmd(bin, Args("remote", "add", "origin", g.repo), Chdir(cfg.Dir)), rt); res.Err != nil {
			return res.Err
		}
		if res := Run(NewCmd(bin, Args("add", "."), Chdir(cfg.Dir)), rt); res.Err != nil {
			return res.Err
		}
		if res := Run(NewCmd(bin, Args("commit", "-m", "Import"), Chdir(cfg.Dir)), rt); res.Err != nil {
			return res.Err
		}
		if res := Run(NewCmd(bin, Args("push", "origin", "master"), Chdir(cfg.Dir)), rt); res.Err != nil {
			return res.Err
		}
	} else {
		if res := Run(NewCmd(bin, Args("init"), Chdir(cfg.Dir)), rt); res.Err != nil {
			return res.Err
		}
		if res := Run(NewCmd(bin, Args("remote", "add", "origin", g.repo), Chdir(cfg.Dir)), rt); res.Err != nil {
			return res.Err
		}
	}
	return nil
}

// Sync commands.
func (g *Git) Sync(cfg Config, opt ...SyncOption) error {
	opts := newSyncOptions(opt...)
	rt := opts.Runtime
	if cfg.Dir == "" {
		return errors.Errorf("no sync dir")
	}

	exists, err := pathExists(filepath.Join(cfg.Dir, ".git"))
	if err != nil {
		return err
	}
	if !exists {
		if err := g.setup(cfg, rt); err != nil {
			return err
		}
	}

	if err := checkIgnore(cfg); err != nil {
		return err
	}

	bin, err := exec.LookPath("git")
	if err != nil {
		return err
	}

	status := Run(NewCmd(bin, Args("status", "--porcelain"), Chdir(cfg.Dir)), rt)
	if status.Err != nil {
		return status.Err
	}
	changes := string(status.Output.Stdout)
	if changes != "" {
		if res := Run(NewCmd(bin, Args("add", "."), Chdir(cfg.Dir)), rt); res.Err != nil {
			return res.Err
		}
		if res := Run(NewCmd(bin, Args("commit", "-m", "Sync"), Chdir(cfg.Dir)), rt); res.Err != nil {
			return res.Err
		}
		if res := Run(NewCmd(bin, Args("pull", "--rebase", "origin", "master"), Chdir(cfg.Dir)), rt); res.Err != nil {
			return res.Err
		}
		if res := Run(NewCmd(bin, Args("push", "origin", "master"), Chdir(cfg.Dir)), rt); res.Err != nil {
			return res.Err
		}
	} else {
		if res := Run(NewCmd(bin, Args("pull", "--rebase", "origin", "master"), Chdir(cfg.Dir)), rt); res.Err != nil {
			return res.Err
		}
	}
	return nil
}

// Clean for gsutil is a noop.
func (g *Git) Clean(cfg Config) error {
	if err := removeIfExists(filepath.Join(cfg.Dir, ".git")); err != nil {
		return err
	}
	if err := removeIfExists(filepath.Join(cfg.Dir, ".gitignore")); err != nil {
		return err
	}
	return nil
}

func removeIfExists(path string) error {
	exists, err := pathExists(path)
	if err != nil {
		return err
	}
	if exists {
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}
	return nil
}

func ensureIgnore(dir string, rt Runtime) error {
	path := filepath.Join(dir, ".gitignore")
	exists, err := pathExists(path)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	rt.Log("Creating %s", path)
	b := []byte(".*\n!/.gitignore\n")
	if err := ioutil.WriteFile(path, b, 0600); err != nil {
		return err
	}
	return nil
}

// checkIgnore checks to see that we either have a .gitignore present or that no
// dot (.) files exist.
// This check is necessary when first setting up the repo, since we can't import
// files to a non-empty repo and skip existing hidden files at the same time
// (without stashing) since we need to pull the .gitignore first.
func checkIgnore(cfg Config) error {
	path := filepath.Join(cfg.Dir, ".gitignore")
	exists, err := pathExists(path)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	files, err := ioutil.ReadDir(cfg.Dir)
	if err != nil {
		return nil
	}
	for _, f := range files {
		if f.Name() == ".git" {
			continue
		}
		if strings.HasPrefix(f.Name(), ".") {
			return errors.Errorf("hidden file exists before we can initialize .gitignore: %s", f.Name())
		}
	}
	return nil
}
