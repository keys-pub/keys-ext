package syncp

import (
	"bytes"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// Program to run for syncing.
type Program interface {
	// Sync runs the sync commands.
	Sync(Config, ...SyncOption) error

	// Clean removes files created by program.
	Clean(Config) error
}

// NewProgram creates a program.
func NewProgram(name string, remote string) (Program, error) {
	switch name {
	case "gsutil":
		return NewGSUtil(remote)
	case "git":
		return NewGit(remote)
	case "awss3":
		return NewAWSS3(remote)
	case "rclone":
		return NewRClone(remote)
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

// Runtime keeps information about the running program, like logs.
type Runtime interface {
	// Log ...
	Log(format string, args ...interface{})
}

type runtime struct{}

func (r runtime) Log(format string, args ...interface{}) {}

// Run a command.
func Run(c Cmd, rt Runtime) Result {
	if rt == nil {
		rt = runtime{}
	}
	result := Result{Cmd: c}
	if c.Opts.Chdir != "" {
		if err := os.Chdir(c.Opts.Chdir); err != nil {
			result.Err = errors.Wrapf(err, "failed to chdir")
			return result
		}
	}

	rt.Log("%s %v", c.Bin, c.Opts.Args)
	// TODO: This could be dangerous in a privileged environment
	cmd := exec.Command(c.Bin, c.Opts.Args...) // #nosec

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	result.Err = cmd.Run()
	result.Output = Output{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}

	out := result.Output.String()
	if out != "" {
		rt.Log("%s", result.Output.String())
	}

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
