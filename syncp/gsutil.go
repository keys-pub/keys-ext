package syncp

import (
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// GSUtil describes how to run gsutil.
type GSUtil struct {
	bucket string
}

// NewGSUtil creates gcp storage rsync command.
func NewGSUtil(bucket string) (*GSUtil, error) {
	if err := validateGSBucket(bucket); err != nil {
		return nil, err
	}
	return &GSUtil{
		bucket: bucket,
	}, nil
}

func validateGSBucket(s string) error {
	if s == "" {
		return errors.Errorf("no gsutil bucket specified")
	}
	if !strings.HasPrefix(s, "gs://") {
		return errors.Errorf("invalid bucket scheme, expected gs://")
	}
	return nil
}

func (g *GSUtil) pushArgs(cfg Config) []string {
	args := []string{
		"-m",
		"rsync",
		"-e", // Exclude symlinks
		"-x", // Exclude .git
		`\..*`,
		cfg.Dir,
		g.bucket,
	}
	return args
}

func (g *GSUtil) pullArgs(cfg Config) []string {
	args := []string{
		"-m",
		"rsync",
		"-e", // Exclude symlinks
		"-x", // Exclude .git
		`\..*`,
		g.bucket,
		cfg.Dir,
	}
	return args
}

// Sync commands.
func (g *GSUtil) Sync(cfg Config, opt ...SyncOption) error {
	opts := newSyncOptions(opt...)
	rt := opts.Runtime
	bin, err := exec.LookPath("gsutil")
	if err != nil {
		return err
	}

	if res := Run(NewCmd(bin, Args(g.pushArgs(cfg)...)), rt); res.Err != nil {
		return res.Err
	}
	if res := Run(NewCmd(bin, Args(g.pullArgs(cfg)...)), rt); res.Err != nil {
		return res.Err
	}

	return nil
}

// Clean for gsutil is a noop.
func (g *GSUtil) Clean(cfg Config) error {
	return nil
}
