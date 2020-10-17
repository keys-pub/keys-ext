package fido2

import (
	"plugin"

	"github.com/pkg/errors"
)

// OpenPlugin returns AuthServer from shared library.
func OpenPlugin(path string) (FIDO2Server, error) {
	plug, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}
	symLibrary, err := plug.Lookup("FIDO2Server")
	if err != nil {
		return nil, err
	}

	lib, ok := symLibrary.(FIDO2Server)
	if !ok {
		return nil, errors.Errorf("not FIDO2Server library")
	}
	return lib, nil
}
