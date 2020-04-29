package fido2

import (
	"plugin"

	"github.com/pkg/errors"
)

// OpenPlugin returns AuthenticatorsServer from shared library.
func OpenPlugin(path string) (AuthenticatorsServer, error) {
	plug, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}
	symLibrary, err := plug.Lookup("AuthenticatorsServer")
	if err != nil {
		return nil, err
	}

	lib, ok := symLibrary.(AuthenticatorsServer)
	if !ok {
		return nil, errors.Errorf("not AuthenticatorsServer library")
	}
	return lib, nil
}
