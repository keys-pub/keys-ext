package sync

import "os/exec"

// Git describes how to run git syncing.
type Git struct{}

// NewGit creates git syncing command.
func NewGit() (*Git, error) {
	return &Git{}, nil
}

// Commands to be run.
func (g *Git) Commands(cfg Config) ([]Command, error) {
	binPath, err := exec.LookPath("git")
	if err != nil {
		return nil, err
	}
	return []Command{
		Command{
			BinPath: binPath,
			Args: []string{
				"pull",
				"origin",
				"master",
			},
			Chdir: cfg.Dir,
		},
		Command{
			BinPath: binPath,
			Args: []string{
				"add",
				".",
			},
			Chdir: cfg.Dir,
		},
		Command{
			BinPath: binPath,
			Args: []string{
				"commit",
				"-m",
				"Syncing...",
			},
			Chdir: cfg.Dir,
		},
		Command{
			BinPath: binPath,
			Args: []string{
				"push",
				"origin",
				"master",
			},
			Chdir: cfg.Dir,
		},
	}, nil
}

// GitSetup describes how to setup a directory with git.
//
// To setup, initialize the git repo and add the remote:
// git init
// git remote add origin git@gitlab.com:gabrielha/keys.pub.test.git
type GitSetup struct {
	repo string
}

// NewGitSetup creates git sync command.
func NewGitSetup(repo string) (*GitSetup, error) {
	return &GitSetup{
		repo: repo,
	}, nil
}

// Commands to be run.
func (g *GitSetup) Commands(cfg Config) ([]Command, error) {
	binPath, err := exec.LookPath("git")
	if err != nil {
		return nil, err
	}
	return []Command{
		Command{
			BinPath: binPath,
			Args: []string{
				"init",
			},
			Chdir: cfg.Dir,
		},
		Command{
			BinPath: binPath,
			Args: []string{
				"remote", "add", "origin", g.repo,
			},
			Chdir: cfg.Dir,
		},
	}, nil
}
