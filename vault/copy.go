package vault

import "github.com/pkg/errors"

// Copy data from a vault.Store to vault vault.Store.
// It copies raw data, it doesn't need to be unlocked.
func Copy(from Store, to Store, opt ...CopyOption) ([]string, error) {
	opts := newCopyOptions(opt...)

	docs, err := from.Documents()
	if err != nil {
		return nil, err
	}

	added := []string{}
	for _, doc := range docs {
		path, b := doc.Path, doc.Data
		data, err := to.Get(path)
		if err != nil {
			return nil, err
		}
		if data != nil {
			if opts.SkipExisting {
				continue
			} else {
				return nil, errors.Errorf("failed to copy: already exists %s", path)
			}
		}
		if !opts.DryRun {
			if err := to.Set(path, b); err != nil {
				return nil, err
			}
		}
		added = append(added, path)
	}

	return added, nil
}

// CopyOption ...
type CopyOption func(*CopyOptions)

// CopyOptions ...
type CopyOptions struct {
	SkipExisting bool
	DryRun       bool
}

func newCopyOptions(opts ...CopyOption) CopyOptions {
	var options CopyOptions
	for _, o := range opts {
		o(&options)
	}
	return options
}

// SkipExisting to skip existing entries, otherwise error.
func SkipExisting() CopyOption {
	return func(o *CopyOptions) {
		o.SkipExisting = true
	}
}

// DryRun to pretend to copy.
func DryRun() CopyOption {
	return func(o *CopyOptions) {
		o.DryRun = true
	}
}
