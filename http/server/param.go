package server

import (
	"fmt"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func queryParamInt(c echo.Context, name string, dft int) (int, error) {
	param := c.QueryParam(name)
	if param == "" {
		return dft, nil
	}
	i, err := strconv.Atoi(param)
	if err != nil {
		return 0, errors.Wrapf(err, fmt.Sprintf("invalid %s", name))
	}
	return i, nil
}

func queryParamDuration(c echo.Context, name string, dft time.Duration) (time.Duration, error) {
	param := c.QueryParam(name)
	if param == "" {
		return dft, nil
	}
	d, err := time.ParseDuration(param)
	if err != nil {
		return 0, errors.Wrapf(err, fmt.Sprintf("invalid %s", name))
	}
	return d, nil
}
