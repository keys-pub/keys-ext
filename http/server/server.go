package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/keys-pub/keys"
	"github.com/labstack/echo/v4"

	"github.com/pkg/errors"
)

// TODO: Support If-Modified-Since

// TODO: Turn off logging

// Server ...
type Server struct {
	fi     Fire
	mc     MemCache
	nowFn  func() time.Time
	logger Logger

	// URL (base) of form http(s)://host:port with no trailing slash to help
	// authorization checks in testing where the host is ambiguous.
	URL string

	accessFn AccessFn

	users        *keys.UserStore
	tasks        Tasks
	internalAuth string
}

// Fire defines interface for remote store (like Firestore).
type Fire interface {
	keys.DocumentStore
	keys.Changes
}

// NewServer creates a Server.
func NewServer(fi Fire, mc MemCache, users *keys.UserStore, logger Logger) *Server {
	return &Server{
		fi:     fi,
		mc:     mc,
		nowFn:  time.Now,
		tasks:  newUnsetTasks(),
		users:  users,
		logger: logger,
		accessFn: func(c AccessContext, resource AccessResource, action AccessAction) Access {
			return AccessDeny("no access set")
		},
	}
}

// SetInternalAuth for authorizing internal requests, like tasks.
func (s *Server) SetInternalAuth(internalAuth string) {
	s.internalAuth = internalAuth
}

// SetTasks ...
func (s *Server) SetTasks(tasks Tasks) {
	s.tasks = tasks
}

// NewHandler returns http.Handler for Server.
func NewHandler(s *Server) http.Handler {
	return newHandler(s)
}

func newHandler(s *Server) *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = ErrorHandler
	s.AddRoutes(e)
	return e
}

// AddRoutes adds routes to an Echo instance.
func (s *Server) AddRoutes(e *echo.Echo) {
	e.GET("/sigchain/:kid/:seq", s.getSigchainStatement)
	e.PUT("/sigchain/:kid/:seq", s.putSigchainStatement)
	e.GET("/sigchain/:kid", s.getSigchain)

	e.POST("/check", s.check)

	e.GET("/user/search", s.getUserSearch)
	e.GET("/user/:kid", s.getUser)

	// Tasks
	e.POST("/task/check/:kid", s.taskCheck)
	e.POST("/task/expired", s.taskExpired)
	e.GET("/task/create/check/:kid", s.createTaskCheck)

	// Cron
	e.POST("/cron/check", s.cronCheck)
	e.POST("/cron/expired", s.cronExpired)

	// Messages
	e.POST("/msgs/:kid/:rid", s.postMessage)
	e.GET("/msgs/:kid/:rid", s.listMessages)

	// Disco
	e.PUT("/disco/:kid/:rid/:type", s.putDisco)
	e.GET("/disco/:kid/:rid/:type", s.getDisco)
	e.DELETE("/disco/:kid/:rid", s.deleteDisco)

	// Invite
	e.POST("/invite/:kid/:rid", s.postInvite)
	e.GET("/invite", s.getInvite)

	// Sigchain (aliases)
	e.GET("/:kid", s.getSigchain)
	e.GET("/:kid/:seq", s.getSigchainStatement)
	e.PUT("/:kid/:seq", s.putSigchainStatement)
}

// SetNowFn sets clock Now function.
func (s *Server) SetNowFn(nowFn func() time.Time) {
	s.nowFn = nowFn
}

// JSON response.
func JSON(c echo.Context, status int, i interface{}) error {
	var b []byte
	switch v := i.(type) {
	case []byte:
		b = v
	default:
		mb, err := json.Marshal(i)
		if err != nil {
			panic(err)
		}
		b = mb
	}
	return c.Blob(status, echo.MIMEApplicationJSONCharsetUTF8, b)
}

func (s *Server) checkInternalAuth(c echo.Context) error {
	if s.internalAuth == "" {
		return ErrForbidden(c, errors.Errorf("no auth token set on server"))
	}
	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		return ErrForbidden(c, errors.Errorf("no auth token specified"))
	}
	if auth != s.internalAuth {
		return ErrForbidden(c, errors.Errorf("invalid auth token"))
	}
	return nil
}
