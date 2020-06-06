package sync

import (
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// Program to run for syncing.
type Program interface {
	Commands(cfg Config) ([]Command, error)
}

// Command describes the commands that are run.
type Command struct {
	BinPath string
	Args    []string
	Chdir   string
}

// Config describes the current runtime environment.
type Config struct {
	Dir string
}

// Run is the default program run/exec.
func Run(pr Program, cfg Config) error {
	if cfg.Dir == "" {
		return errors.Errorf("invalid sync dir: %q", cfg.Dir)
	}

	cmds, err := pr.Commands(cfg)
	if err != nil {
		return err
	}

	for _, c := range cmds {
		if c.Chdir != "" {
			if err := os.Chdir(c.Chdir); err != nil {
				return errors.Wrapf(err, "failed to chdir (sync)")
			}
		}

		// TODO: If this ever runs under a privileged environment we need to be
		// careful that the PATH only includes privileged locations.
		cmd := exec.Command(c.BinPath, c.Args...) // #nosec
		logger.Infof("Running %s %s", c.BinPath, c.Args)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return errors.Wrapf(err, "failed to run sync command %s %v; %s", c.BinPath, c.Args, replaceNewlines(string(out)))
		}
	}
	return nil
}

func replaceNewlines(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\r\n", "; ")
	s = strings.ReplaceAll(s, "\n", "; ")
	return s
}
