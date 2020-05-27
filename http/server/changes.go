package server

import (
	"strconv"

	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/tsutil"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type changes struct {
	docs          []*ds.Document
	version       int64
	versionNext   int64
	errBadRequest error
}

func (s *Server) changes(c echo.Context, path string) (*changes, error) {
	request := c.Request()
	ctx := request.Context()

	var version int64
	if f := c.QueryParam("version"); f != "" {
		i, err := strconv.Atoi(f)
		if err != nil {
			return &changes{errBadRequest: errors.Wrapf(err, "invalid version")}, nil
		}
		version = int64(i)
	}
	plimit := c.QueryParam("limit")
	if plimit == "" {
		plimit = "100"
	}
	limit, err := strconv.Atoi(plimit)
	if err != nil {
		return &changes{errBadRequest: errors.Wrapf(err, "invalid limit")}, nil
	}
	if limit > 100 {
		return &changes{errBadRequest: errors.Errorf("invalid limit, too large")}, nil
	}

	pdir := c.QueryParam("direction")
	if pdir == "" {
		pdir = "asc"
	}

	var dir ds.Direction
	switch pdir {
	case "asc":
		dir = ds.Ascending
	case "desc":
		dir = ds.Descending
	default:
		return &changes{errBadRequest: errors.Errorf("invalid dir")}, nil
	}

	s.logger.Infof("Changes %s", path)
	chngs, to, err := s.fi.Changes(ctx, path, tsutil.ParseMillis(version), limit, dir)
	if err != nil {
		return nil, err
	}

	s.logger.Infof("Changes %s, found %d", path, len(chngs))
	paths := make([]string, 0, len(chngs))
	for _, a := range chngs {
		paths = append(paths, a.Path)
	}
	out, err := s.fi.GetAll(ctx, paths)
	if err != nil {
		return nil, err
	}
	s.logger.Debugf("Changes %s, got docs %d", path, len(out))

	versionNext := int64(0)
	if to.IsZero() {
		versionNext = version
	} else {
		versionNext = tsutil.Millis(to)
	}

	s.logger.Infof("Changes %s, version next: %d", path, versionNext)

	return &changes{
		docs:        out,
		version:     int64(version),
		versionNext: int64(versionNext),
	}, nil
}
