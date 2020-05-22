package fido2

import (
	"plugin"

	"github.com/pkg/errors"
)

// OpenPlugin returns AuthsServer from shared library.
func OpenPlugin(path string) (AuthServer, error) {
	plug, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}
	symLibrary, err := plug.Lookup("AuthServer")
	if err != nil {
		return nil, err
	}

	lib, ok := symLibrary.(AuthServer)
	if !ok {
		return nil, errors.Errorf("not AuthServer library")
	}
	return lib, nil
}
