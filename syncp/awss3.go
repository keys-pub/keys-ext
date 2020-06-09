package syncp

import (
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// AWSS3 describes how to run aws s3.
type AWSS3 struct {
	bucket string
}

// NewAWSS3 creates aws sync command.
func NewAWSS3(bucket string) (*AWSS3, error) {
	if err := validateS3Bucket(bucket); err != nil {
		return nil, err
	}
	return &AWSS3{
		bucket: bucket,
	}, nil
}

func validateS3Bucket(s string) error {
	if s == "" {
		return errors.Errorf("no s3 bucket specified")
	}
	if !strings.HasPrefix(s, "s3://") {
		return errors.Errorf("invalid bucket scheme, expected s3://")
	}
	return nil
}

func (a *AWSS3) pushArgs(cfg Config) []string {
	args := []string{
		"s3",
		"sync",
		"--exclude",
		`.*`,
		cfg.Dir,
		a.bucket,
	}
	return args
}

func (a *AWSS3) pullArgs(cfg Config) []string {
	args := []string{
		"s3",
		"sync",
		"--exclude",
		`.*`,
		a.bucket,
		cfg.Dir,
	}
	return args
}

// Sync commands.
func (a *AWSS3) Sync(cfg Config, opt ...SyncOption) error {
	opts := newSyncOptions(opt...)
	bin, err := exec.LookPath("aws")
	if err != nil {
		return err
	}

	if res := Run(NewCmd(bin, Args(a.pushArgs(cfg)...)), opts.Runtime); res.Err != nil {
		return res.Err
	}
	if res := Run(NewCmd(bin, Args(a.pullArgs(cfg)...)), opts.Runtime); res.Err != nil {
		return res.Err
	}

	return nil
}

// Clean for aws s3 is a noop.
func (a *AWSS3) Clean(cfg Config) error {
	return nil
}
