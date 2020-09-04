package service

import (
	"context"

	"github.com/pkg/errors"
)

func (*service) ConfigGet(ctx *context.Context, req *ConfigGetRequest) (*ConfigGetResponse, error) {
	return nil, errors.Errorf("not implemented")
}

func (*service) ConfigSet(ctx *context.Context, req *ConfigSetRequest) (*ConfigSetResponse, error) {
	return nil, errors.Errorf("not implemented")
}
