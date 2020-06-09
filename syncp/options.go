package syncp

// SyncOptions ...
type SyncOptions struct {
	Runtime Runtime
}

// SyncOption ...
type SyncOption func(*SyncOptions)

func newSyncOptions(opt ...SyncOption) SyncOptions {
	opts := SyncOptions{
		Runtime: runtime{},
	}
	for _, o := range opt {
		o(&opts)
	}
	return opts
}

// WithRuntime ..
func WithRuntime(rt Runtime) SyncOption {
	return func(o *SyncOptions) {
		o.Runtime = rt
	}
}

// CmdOptions ...
type CmdOptions struct {
	Args  []string
	Chdir string
}

// CmdOption ...
type CmdOption func(*CmdOptions)

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
