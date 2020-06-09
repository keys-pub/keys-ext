package syncp

import (
	"os/exec"

	"github.com/pkg/errors"
)

// RClone describes how to run rclone.
type RClone struct {
	remote string
}

// NewRClone creates rclone sync command.
func NewRClone(remote string) (*RClone, error) {
	if remote == "" {
		return nil, errors.Errorf("no remote specified")
	}
	return &RClone{
		remote: remote,
	}, nil
}

func (r *RClone) pushArgs(cfg Config) []string {
	args := []string{
		"copy",
		"--exclude",
		`\.*`,
		cfg.Dir,
		r.remote,
	}
	return args
}

func (r *RClone) pullArgs(cfg Config) []string {
	args := []string{
		"copy",
		"--exclude",
		`\.*`,
		r.remote,
		cfg.Dir,
	}
	return args
}

// Sync commands.
func (r *RClone) Sync(cfg Config, opt ...SyncOption) error {
	opts := newSyncOptions(opt...)
	rt := opts.Runtime
	bin, err := exec.LookPath("rclone")
	if err != nil {
		return err
	}

	if res := Run(NewCmd(bin, Args(r.pushArgs(cfg)...)), rt); res.Err != nil {
		return res.Err
	}
	if res := Run(NewCmd(bin, Args(r.pullArgs(cfg)...)), rt); res.Err != nil {
		return res.Err
	}

	return nil
}

// Clean for rclone is a noop.
func (r *RClone) Clean(cfg Config) error {
	return nil
}
