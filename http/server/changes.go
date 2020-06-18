package server

import (
	"strconv"

	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/tsutil"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type changes struct {
	changes     []*ds.Change
	version     int64
	versionNext int64
}

func (s *Server) changes(c echo.Context, path string) (*changes, error, error) {
	request := c.Request()
	ctx := request.Context()

	var version int64
	if f := c.QueryParam("version"); f != "" {
		i, err := strconv.Atoi(f)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid version"), nil
		}
		version = int64(i)
	}
	plimit := c.QueryParam("limit")
	if plimit == "" {
		plimit = "100"
	}
	limit, err := strconv.Atoi(plimit)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid limit"), nil
	}
	if limit > 100 {
		return nil, errors.Errorf("invalid limit, too large"), nil
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
		return nil, errors.Errorf("invalid dir"), nil
	}

	s.logger.Infof("Changes %s (from=%d)", path, version)
	iter, err := s.fi.Changes(ctx, path, tsutil.ParseMillis(version), limit, dir)
	if err != nil {
		return nil, nil, err
	}
	defer iter.Release()
	chngs, to, err := ds.ChangesFromIterator(iter, tsutil.ParseMillis(version))
	if err != nil {
		return nil, nil, err
	}

	s.logger.Debugf("Changes %s, got %d", path, len(chngs))

	versionNext := int64(0)
	if to.IsZero() {
		versionNext = version
	} else {
		versionNext = tsutil.Millis(to)
	}

	s.logger.Infof("Changes %s (to=%d)", path, versionNext)

	return &changes{
		changes:     chngs,
		version:     int64(version),
		versionNext: int64(versionNext),
	}, nil, nil
}
