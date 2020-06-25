package server

import (
	"io/ioutil"
	"net/http"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/ds"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (s *Server) postMessage(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	kid, status, err := authorize(c, s.URL, "kid", s.nowFn(), s.rds)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	recipient := c.Param("rid")
	if recipient == "" {
		return ErrBadRequest(c, errors.Errorf("no recipient id"))
	}
	rid, err := keys.ParseID(recipient)
	if err != nil {
		return ErrBadRequest(c, err)
	}

	// expire := time.Hour * 24
	// if c.QueryParam("expire") != "" {
	// 	e, err := time.ParseDuration(c.QueryParam("expire"))
	// 	if err != nil {
	// 		return ErrBadRequest(c, err)
	// 	}
	// 	expire = e
	// }
	// if len(expire.String()) > 64 {
	// 	return ErrBadRequest(c, errors.Errorf("invalid expire"))
	// }
	// if expire > time.Hour*24 {
	// 	return ErrBadRequest(c, errors.Errorf("max expire is 24h"))
	// }
	// exp := expire.String()
	// if expire == time.Duration(0) {
	// 	exp = ""
	// }

	if c.Request().Body == nil {
		return ErrBadRequest(c, errors.Errorf("missing body"))
	}

	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return s.internalError(c, err)
	}

	if len(b) > 16*1024 {
		// TODO: Check length before reading data
		return ErrBadRequest(c, errors.Errorf("message too large (greater than 16KiB)"))
	}

	ctx := c.Request().Context()

	var path string
	if kid != rid {
		addr, err := keys.NewAddress(kid, rid)
		if err != nil {
			return ErrBadRequest(c, errors.Wrapf(err, "invalid address"))
		}
		path = ds.Path("msgs", addr)
	} else {
		path = ds.Path("msgs", kid)
	}

	events, err := s.fi.EventsAdd(ctx, path, [][]byte{b})
	if err != nil {
		return s.internalError(c, err)
	}
	if len(events) == 0 {
		return s.internalError(c, errors.Errorf("no events added"))
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) listMessages(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	kid, status, err := authorize(c, s.URL, "kid", s.nowFn(), s.rds)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	recipient := c.Param("rid")
	if recipient == "" {
		return ErrBadRequest(c, errors.Errorf("no recipient id"))
	}
	rid, err := keys.ParseID(recipient)
	if err != nil {
		return ErrBadRequest(c, err)
	}

	var path string
	if kid != rid {
		addr, err := keys.NewAddress(kid, rid)
		if err != nil {
			return ErrBadRequest(c, errors.Wrapf(err, "invalid address"))
		}
		path = ds.Path("msgs", addr)
	} else {
		path = ds.Path("msgs", kid)
	}

	resp, respErr := s.events(c, path)
	if respErr != nil {
		return respErr
	}

	return JSON(c, http.StatusOK, resp)
}

// func (s *Server) checkMessage(msg *message, doc *ds.Document) (bool, error) {
// 	if msg.Expire != "" {
// 		expiry, err := time.ParseDuration(msg.Expire)
// 		if err != nil {
// 			return false, errors.Wrapf(err, "invalid expire")
// 		}
// 		now := s.nowFn()
// 		s.logger.Debugf("Now: %s, Created: %s, Expiry: %s, Sub: %s", now, doc.CreatedAt, expiry, now.Sub(doc.CreatedAt))
// 		if now.Sub(doc.CreatedAt) > expiry {
// 			return false, nil
// 		}
// 	}
// 	return true, nil
// }

// func truthy(s string) bool {
// 	s = strings.TrimSpace(s)
// 	switch s {
// 	case "", "0", "f", "false", "n", "no":
// 		return false
// 	default:
// 		return true
// 	}
// }
