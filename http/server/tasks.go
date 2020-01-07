package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/keys-pub/keys"
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
		return internalError(c, err)
	}
	return c.String(http.StatusOK, "")
}

func (s *Server) taskCheck(c echo.Context) error {
	ctx := c.Request().Context()
	logger.Infof(ctx, "Server POST check %s", s.urlString(c))

	if err := s.checkInternalAuth(c); err != nil {
		return err
	}

	kid, err := keys.ParseID(c.Param("kid"))
	if err != nil {
		return ErrBadRequest(c, err)
	}

	if _, err := s.users.Update(ctx, kid); err != nil {
		return internalError(c, err)
	}
	return c.String(http.StatusOK, "")
}

func (s *Server) cronCheck(c echo.Context) error {
	ctx := c.Request().Context()
	logger.Infof(ctx, "Server POST cron check %s", s.urlString(c))

	kids, err := s.users.Expired(ctx, time.Hour*23)
	if err != nil {
		return internalError(c, err)
	}

	// logger.Infof(ctx, "Expired %s", kids)

	for _, kid := range kids {
		if err := s.tasks.CreateTask(ctx, "POST", "/task/check/"+kid.String(), s.internalAuth); err != nil {
			return internalError(c, err)
		}
	}

	return c.String(http.StatusOK, "")
}
