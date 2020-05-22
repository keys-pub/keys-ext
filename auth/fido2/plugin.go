package fido2

import (
	"plugin"

	"github.com/pkg/errors"
)

// OpenPlugin returns AuthsServer from shared library.
func OpenPlugin(path string) (AuthsServer, error) {
	plug, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}
	symLibrary, err := plug.Lookup("AuthsServer")
	if err != nil {
		return nil, err
	}

	lib, ok := symLibrary.(AuthsServer)
	if !ok {
		return nil, errors.Errorf("not AuthsServer library")
	}
	return lib, nil
}
