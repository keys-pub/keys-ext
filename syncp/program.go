package syncp

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// Program to run for syncing.
type Program interface {
	// Sync runs the sync commands.
	Sync(cfg Config, rt Runtime) error

	// Setup runs setup commands.
	Setup(cfg Config, rt Runtime) error

	// Clean runs cleanup commands.
	Clean(cfg Config, rt Runtime) error
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
	Bin  string
	Opts CmdOptions
}

// NewCmd creates a command.
func NewCmd(bin string, opt ...CmdOption) Cmd {
	opts := newCmdOptions(opt...)
	return Cmd{
		Bin:  bin,
		Opts: opts,
	}
}

// Config describes the current runtime environment.
type Config struct {
	Dir string
}

// Runtime ...
type Runtime interface {
	Log(format string, args ...interface{})
	Logs() []string
}

type runtime struct {
	logs []string
}

// NewRuntime creates a Log.
func NewRuntime() Runtime {
	return &runtime{}
}

// Log ...
func (l *runtime) Log(format string, args ...interface{}) {
	l.logs = append(l.logs, fmt.Sprintf(format, args...))
}

func (l *runtime) Logs() []string {
	return l.logs
}

// Run a command.
func Run(c Cmd, rt Runtime) Result {
	result := Result{Cmd: c}
	if c.Opts.Chdir != "" {
		if err := os.Chdir(c.Opts.Chdir); err != nil {
			result.Err = errors.Wrapf(err, "failed to chdir")
			return result
		}
	}

	// TODO: This could be dangerous in a privileged environment
	cmd := exec.Command(c.Bin, c.Opts.Args...) // #nosec
	logger.Infof("Running %s %s", c.Bin, c.Opts.Args)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	result.Err = cmd.Run()
	result.Output = Output{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}

	rt.Log("%s %v", c.Bin, c.Opts.Args)
	rt.Log("%s", result.Output.String())

	return result
}

// Result from running command.
type Result struct {
	Cmd    Cmd
	Output Output
	Err    error
}

// Output from running command.
type Output struct {
	Stdout []byte
	Stderr []byte
}

func (o Output) String() string {
	s := []string{}
	out := strings.TrimSpace(string(o.Stdout))
	if out != "" {
		s = append(s, out)
	}
	err := strings.TrimSpace(string(o.Stderr))
	if err != "" {
		s = append(s, err)
	}
	return strings.Join(s, "\n")
}

// ExitCode returns 0 if no error, the exit code if an exec.ExitError, or -1 if
// a different error.
func (r Result) ExitCode() int {
	if r.Err == nil {
		return 0
	}
	if exitError, ok := r.Err.(*exec.ExitError); ok {
		return exitError.ExitCode()
	}
	return -1
}

func pathExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}
