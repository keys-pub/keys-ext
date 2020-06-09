package syncp

// CmdOptions ...
type CmdOptions struct {
	Args  []string
	Chdir string
}

// CmdOption ...
type CmdOption func(*CmdOptions)

// NewCmd creates a Cmd.
func newCmdOptions(opt ...CmdOption) CmdOptions {
	var opts CmdOptions
	for _, o := range opt {
		o(&opts)
	}
	return opts
}

// Args ..
func Args(args ...string) CmdOption {
	return func(o *CmdOptions) {
		o.Args = args
	}
}

// Chdir ..
func Chdir(dir string) CmdOption {
	return func(o *CmdOptions) {
		o.Chdir = dir
	}
}
