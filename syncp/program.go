package syncp

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// Program to run for syncing.
type Program interface {
	// Setup runs optional setup commands.
	Setup(cfg Config) Result
	// Sync runs the sync commands.
	Sync(cfg Config) Result
}

// NewProgram creates a program.
func NewProgram(name string, remote string) (Program, error) {
	switch name {
	case "gsutil":
		return NewGSUtil(remote)
	case "git":
		return NewGit(remote)
	default:
		return nil, errors.Errorf("unrecognized sync program name %q", name)
	}
}

// Cmd describes the commands that are run.
type Cmd struct {
	BinPath string
	Args    []string
	Chdir   string
}

// Result from running commands.
type Result struct {
	CmdResults []CmdResult
	Err        error
}

func (r Result) String() string {
	out := []string{}
	for _, r := range r.CmdResults {
		out = append(out, fmt.Sprintf("----"))
		out = append(out, fmt.Sprintf("cmd: %s %v", r.Cmd.BinPath, r.Cmd.Args))
		out = append(out, fmt.Sprintf("out: %s", string(r.Output)))
		out = append(out, fmt.Sprintf("exc: %d", r.ExitCode()))
		if r.Err != nil {
			out = append(out, fmt.Sprintf("err: %v", r.Err))
		}
	}
	return strings.Join(out, "\n")
}

// CmdResult from running command.
type CmdResult struct {
	Cmd    Cmd
	Output []byte
	Err    error
}

// ExitCode returns 0 if no error, the exit code if an exec.ExitError, or -1 if
// a different error.
func (r CmdResult) ExitCode() int {
	if r.Err == nil {
		return 0
	}
	if exitError, ok := r.Err.(*exec.ExitError); ok {
		return exitError.ExitCode()
	}
	return -1
}

// Config describes the current runtime environment.
type Config struct {
	Dir string
}

// RunAll the commands.
func RunAll(cmds []Cmd, cfg Config) Result {
	result := Result{
		CmdResults: []CmdResult{},
	}
	if cfg.Dir == "" {
		result.Err = errors.Errorf("invalid sync dir: %q", cfg.Dir)
		return result
	}
	for _, c := range cmds {
		cmdResult := Run(c, cfg)
		result.CmdResults = append(result.CmdResults, cmdResult)
		if cmdResult.Err != nil {
			result.Err = cmdResult.Err
			break
		}
	}
	return result
}

// Run a command.
func Run(c Cmd, cfg Config) CmdResult {
	result := CmdResult{Cmd: c}
	if c.Chdir != "" {
		if err := os.Chdir(c.Chdir); err != nil {
			result.Err = errors.Wrapf(err, "failed to chdir (sync)")
			return result
		}
	}

	// TODO: If this ever runs under a privileged environment we need to be
	// careful that the PATH only includes privileged locations.
	cmd := exec.Command(c.BinPath, c.Args...) // #nosec
	logger.Infof("Running %s %s", c.BinPath, c.Args)
	out, err := cmd.CombinedOutput()
	result.Output = out
	if err != nil {
		result.Err = err
	}
	return result
}
