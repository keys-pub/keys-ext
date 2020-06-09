package syncp

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

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

// Setup commands.
func (g *Git) Setup(cfg Config, rt Runtime) error {
	if cfg.Dir == "" {
		return errors.Errorf("no sync dir")
	}

	bin, err := exec.LookPath("git")
	if err != nil {
		return err
	}

	isEmpty := false
	res := Run(NewCmd(bin, Args("ls-remote", g.repo), Chdir(cfg.Dir)), rt)
	if res.Err != nil {
		return res.Err
	}
	refs := string(res.Output.Stdout)
	if refs == "" {
		isEmpty = true
	}

	if isEmpty {
		if err := ensureReadme(cfg.Dir, rt); err != nil {
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
func (g *Git) Sync(cfg Config, rt Runtime) error {
	bin, err := exec.LookPath("git")
	if err != nil {
		return err
	}

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
	return nil
}

// Clean for gsutil is a noop.
func (g *Git) Clean(cfg Config, rt Runtime) error {
	dotGit := filepath.Join(cfg.Dir, ".git")
	exists, err := pathExists(dotGit)
	if err != nil {
		return err
	}
	if exists {
		rt.Log("Removing %s", dotGit)
		if err := os.RemoveAll(dotGit); err != nil {
			return err
		}
	}
	return nil
}

func ensureReadme(dir string, rt Runtime) error {
	path := filepath.Join(dir, "README.md")
	exists, err := pathExists(path)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	rt.Log("Creating %s", path)
	b := []byte("# keys.pub\n")
	if err := ioutil.WriteFile(path, b, 0700); err != nil {
		return err
	}
	return nil
}
