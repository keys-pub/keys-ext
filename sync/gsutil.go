package sync

import (
	"fmt"
	"os/exec"

	"github.com/pkg/errors"
)

// GSUtil describes how to run gsutil.
type GSUtil struct {
	bucketName string
}

// NewGSUtil creates gcp storage rsync command.
//
// To setup, make a remote bucket:
// gsutil mb gs://<bucket>
func NewGSUtil(binPath string, bucketName string) (*GSUtil, error) {
	if bucketName == "" {
		return nil, errors.Errorf("no gsutil bucket name specified")
	}
	return &GSUtil{
		bucketName: bucketName,
	}, nil
}

func (g *GSUtil) pushArgs(cfg Config) []string {
	bucketName := fmt.Sprintf("gs://%s", g.bucketName)
	args := []string{
		"-m",
		"rsync",
		"-e", // Exclude symlinks
		"-x", // Exclude .git
		`\.git$`,
		cfg.Dir,
		bucketName,
	}
	return args
}

func (g *GSUtil) pullArgs(cfg Config) []string {
	bucketName := fmt.Sprintf("gs://%s", g.bucketName)
	args := []string{
		"-m",
		"rsync",
		"-e", // Exclude symlinks
		"-x", // Exclude .git
		`\.git$`,
		bucketName,
		cfg.Dir,
	}
	return args
}

// Commands to be run.
func (g *GSUtil) Commands(cfg Config) ([]Command, error) {
	binPath, err := exec.LookPath("gsutil")
	if err != nil {
		return nil, err
	}
	return []Command{
		Command{
			BinPath: binPath,
			Args:    g.pushArgs(cfg),
		},
		Command{
			BinPath: binPath,
			Args:    g.pullArgs(cfg),
		},
	}, nil
}
