package syncp

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// GSUtil describes how to run gsutil.
type GSUtil struct {
	bucket string
}

// NewGSUtil creates gcp storage rsync command.
//
// To setup, make a remote bucket:
// gsutil mb gs://bucket
func NewGSUtil(bucket string) (*GSUtil, error) {
	if err := validateBucket(bucket); err != nil {
		return nil, err
	}
	return &GSUtil{
		bucket: bucket,
	}, nil
}

func validateBucket(s string) error {
	if s == "" {
		return errors.Errorf("no gsutil bucket specified")
	}
	if !strings.HasPrefix(s, "gs://") {
		return errors.Errorf("invalid bucket scheme, expected gs://")
	}
	return nil
}

// ID returns unique identifier for the program + remote.
func (g *GSUtil) ID() string {
	return fmt.Sprintf("gsutil+" + g.bucket)
}

func (g *GSUtil) String() string {
	return fmt.Sprintf("gsutil+" + g.bucket)
}

func (g *GSUtil) pushArgs(cfg Config) []string {
	args := []string{
		"-m",
		"rsync",
		"-e", // Exclude symlinks
		"-x", // Exclude .git
		`\.git$`,
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
		`\.git$`,
		g.bucket,
		cfg.Dir,
	}
	return args
}

// Setup for gsutil is a noop.
func (g *GSUtil) Setup(cfg Config) Result {
	return Result{}
}

// Sync commands.
func (g *GSUtil) Sync(cfg Config) Result {
	binPath, err := exec.LookPath("gsutil")
	if err != nil {
		return Result{Err: err}
	}
	cmds := []Cmd{
		Cmd{
			BinPath: binPath,
			Args:    g.pushArgs(cfg),
		},
		Cmd{
			BinPath: binPath,
			Args:    g.pullArgs(cfg),
		},
	}
	return RunAll(cmds, cfg)
}
