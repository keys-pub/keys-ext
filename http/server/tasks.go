package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/user"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// Tasks ..
type Tasks interface {
	// CreateTask ...
	CreateTask(ctx context.Context, method string, url string, authToken string) error
}

type testTasks struct {
	svr *Server
}

// NewTestTasks returns Tasks for use in tests.
func NewTestTasks(svr *Server) Tasks {
	return &testTasks{
		svr: svr,
	}
}

func (t testTasks) CreateTask(ctx context.Context, method string, url string, authToken string) error {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", authToken)
	rr := httptest.NewRecorder()
	handler := NewHandler(t.svr)
	handler.ServeHTTP(rr, req)
	if rr.Code != 200 {
		return errors.Errorf("task error %d", rr.Code)
	}
	return nil
}

type noTasks struct{}

func newUnsetTasks() Tasks {
	return &noTasks{}
}

func (t noTasks) CreateTask(ctx context.Context, method string, url string, authToken string) error {
	return errors.Errorf("no server tasks set")
}

func (s *Server) createTaskCheck(c echo.Context) error {
	ctx := c.Request().Context()
	kid, err := keys.ParseID(c.Param("kid"))
	if err != nil {
		return ErrBadRequest(c, err)
	}

	if err := s.tasks.CreateTask(ctx, "POST", "/task/check/"+kid.String(), s.internalAuth); err != nil {
		return s.internalError(c, err)
	}
	return c.String(http.StatusOK, "")
}

func (s *Server) taskCheck(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	if err := s.checkInternalAuth(c); err != nil {
		return err
	}

	kid, err := keys.ParseID(c.Param("kid"))
	if err != nil {
		return ErrBadRequest(c, err)
	}

	if _, err := s.users.Update(ctx, kid); err != nil {
		return s.internalError(c, err)
	}
	return c.String(http.StatusOK, "")
}

func (s *Server) checkUserStatus(ctx context.Context, status user.Status) error {
	kids, err := s.users.Status(ctx, status)
	if err != nil {
		return err
	}
	s.logger.Infof("Checking %s (%d)", status, len(kids))
	if err := s.checkKeys(ctx, kids); err != nil {
		return err
	}
	return nil
}

func (s *Server) checkExpired(ctx context.Context, dt time.Duration, maxAge time.Duration) error {
	kids, err := s.users.Expired(ctx, dt, maxAge)
	if err != nil {
		return err
	}
	s.logger.Infof("Checking expired (%d)", len(kids))
	if err := s.checkKeys(ctx, kids); err != nil {
		return err
	}
	return nil
}

func (s *Server) cronCheck(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	// TODO: Need to test this

	if err := s.checkUserStatus(ctx, user.StatusConnFailure); err != nil {
		return s.internalError(c, err)
	}

	// Check expired
	if err := s.checkExpired(ctx, time.Hour*12, time.Hour*24*60); err != nil {
		return s.internalError(c, err)
	}

	return c.String(http.StatusOK, "")
}

func (s *Server) checkKeys(ctx context.Context, kids []keys.ID) error {
	if len(kids) > 0 {
		s.logger.Infof("Checking %d keys...", len(kids))
	}

	for _, kid := range kids {
		if err := s.tasks.CreateTask(ctx, "POST", "/task/check/"+kid.String(), s.internalAuth); err != nil {
			return err
		}
	}
	return nil
}

// func (s *Server) taskExpired(c echo.Context) error {
// 	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
// 	ctx := c.Request().Context()

// 	if err := s.checkInternalAuth(c); err != nil {
// 		return err
// 	}

// 	iter, err := s.fi.Documents(ctx, ds.Path("msgs"))
// 	if err != nil {
// 		return s.internalError(c, err)
// 	}
// 	defer iter.Release()
// 	paths := []string{}
// 	for {
// 		doc, err := iter.Next()
// 		if err != nil {
// 			return s.internalError(c, err)
// 		}
// 		if doc == nil {
// 			break
// 		}
// 		var msg message
// 		if err := json.Unmarshal(doc.Data, &msg); err != nil {
// 			return s.internalError(c, err)
// 		}

// 		ok, err := s.checkMessage(&msg, doc)
// 		if err != nil {
// 			return s.internalError(c, err)
// 		}
// 		if !ok {
// 			paths = append(paths, doc.Path)
// 		}
// 	}

// 	if err := s.fi.DeleteAll(ctx, paths); err != nil {
// 		return s.internalError(c, err)
// 	}

// 	return c.String(http.StatusOK, "")
// }

func (s *Server) cronExpired(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	if err := s.tasks.CreateTask(ctx, "POST", "/task/expired", s.internalAuth); err != nil {
		return s.internalError(c, err)
	}

	return c.String(http.StatusOK, "")
}
