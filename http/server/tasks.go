package server

import (
	"context"
	"encoding/json"
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

func (s *Server) cronCheck(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	// TODO: Need to test this

	// Check connection failures
	fails, err := s.users.Status(ctx, user.StatusConnFailure)
	if err != nil {
		return s.internalError(c, err)
	}
	if err := s.checkKeys(ctx, fails); err != nil {
		return s.internalError(c, err)
	}

	// Check expired
	kids, err := s.users.Expired(ctx, time.Hour*23)
	if err != nil {
		return s.internalError(c, err)
	}
	if err := s.checkKeys(ctx, kids); err != nil {
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

func (s *Server) taskExpired(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	if err := s.checkInternalAuth(c); err != nil {
		return err
	}

	iter, err := s.fi.Documents(ctx, keys.Path("messages"), nil)
	if err != nil {
		return s.internalError(c, err)
	}
	defer iter.Release()
	paths := []string{}
	for {
		doc, err := iter.Next()
		if err != nil {
			return s.internalError(c, err)
		}
		if doc == nil {
			break
		}
		var msg message
		if err := json.Unmarshal(doc.Data, &msg); err != nil {
			return s.internalError(c, err)
		}

		ok, err := s.checkMessage(&msg, doc)
		if err != nil {
			return s.internalError(c, err)
		}
		if !ok {
			paths = append(paths, doc.Path)
		}
	}

	if err := s.fi.DeleteAll(ctx, paths); err != nil {
		return s.internalError(c, err)
	}

	return c.String(http.StatusOK, "")
}

func (s *Server) cronExpired(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	if err := s.tasks.CreateTask(ctx, "POST", "/task/expired", s.internalAuth); err != nil {
		return s.internalError(c, err)
	}

	return c.String(http.StatusOK, "")
}
