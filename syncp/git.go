package syncp

import (
	"os/exec"
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
// git init
// git remote add origin git@gitlab.com:gabrielha/keys.pub.test.git
func (g *Git) Setup(cfg Config) Result {
	binPath, err := exec.LookPath("git")
	if err != nil {
		return Result{Err: err}
	}
	cmds := []Cmd{
		Cmd{
			BinPath: binPath,
			Args: []string{
				"init",
			},
			Chdir: cfg.Dir,
		},
		Cmd{
			BinPath: binPath,
			Args: []string{
				"remote", "add", "origin", g.repo,
			},
			Chdir: cfg.Dir,
		},
	}
	return RunAll(cmds, cfg)
}

// Sync commands.
func (g *Git) Sync(cfg Config) Result {
	binPath, err := exec.LookPath("git")
	if err != nil {
		return Result{Err: err}
	}
	cmds := []Cmd{
		Cmd{
			BinPath: binPath,
			Args: []string{
				"pull",
				"origin",
				"master",
			},
			Chdir: cfg.Dir,
		},
		Cmd{
			BinPath: binPath,
			Args: []string{
				"add",
				".",
			},
			Chdir: cfg.Dir,
		},
		Cmd{
			BinPath: binPath,
			Args: []string{
				"commit",
				"-m",
				"Syncing...",
			},
			Chdir: cfg.Dir,
		},
		Cmd{
			BinPath: binPath,
			Args: []string{
				"push",
				"origin",
				"master",
			},
			Chdir: cfg.Dir,
		},
	}
	return RunAll(cmds, cfg)
}
