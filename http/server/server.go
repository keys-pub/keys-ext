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

// Server ...
type Server struct {
	fi    Fire
	mc    MemCache
	nowFn func() time.Time

	// URL (base) of form http://host:port with no trailing slash to help
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
func NewServer(fi Fire, mc MemCache, users *keys.UserStore) *Server {
	return &Server{
		fi:    fi,
		mc:    mc,
		nowFn: time.Now,
		tasks: newUnsetTasks(),
		users: users,
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
	AddRoutes(s, e)
	return e
}

// AddRoutes adds routes to an Echo instance.
func AddRoutes(s *Server, e *echo.Echo) {
	e.GET("/sigchain/:kid/:seq", s.getSigchainStatement)
	e.PUT("/sigchain/:kid/:seq", s.putSigchainStatement)
	e.GET("/sigchain/:kid", s.getSigchain)
	e.GET("/sigchains", s.listSigchains)

	e.POST("/check", s.check)

	e.GET("/search", s.getSearch)

	e.PUT("/share/:recipient/:kid", s.putShare)
	e.GET("/share/:recipient/:kid", s.getShare)
	e.DELETE("/share/:recipient/:kid", s.deleteShare)

	e.PUT("/messages/:kid/:id", s.putMessage)
	e.GET("/messages/:kid", s.listMessages)

	e.PUT("/vault/:kid/:id", s.putVault)
	e.GET("/vault/:kid", s.listVault)

	// Tasks
	e.POST("/task/check/:kid", s.taskCheck)
	// Tasks (create)
	e.GET("/task/create/check/:kid", s.createTaskCheck)
	// Cron
	e.POST("/cron/check", s.cronCheck)

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
			return internalError(c, err)
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

func (s *Server) urlString(c echo.Context) string {
	return s.URL + c.Request().URL.String()
}
