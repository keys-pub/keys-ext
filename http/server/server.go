package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/user"
	"github.com/labstack/echo/v4"

	"github.com/pkg/errors"
)

// TODO: Support If-Modified-Since

// TODO: Turn off logging

// Server ...
type Server struct {
	fi     Fire
	rds    Redis
	nowFn  func() time.Time
	logger Logger

	// URL (base) of form http(s)://host:port with no trailing slash to help
	// authorization checks in testing where the host is ambiguous.
	URL string

	accessFn AccessFn

	users        *user.Store
	tasks        Tasks
	internalAuth string

	inc    int64
	incMax int64

	admins []keys.ID
}

// Fire defines interface for remote store (like Firestore).
type Fire interface {
	ds.DocumentStore
	ds.Events
}

// New creates a Server.
func New(fi Fire, rds Redis, users *user.Store, logger Logger) *Server {
	return &Server{
		fi:     fi,
		rds:    rds,
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

// SetAdmins sets authorized admins.
func (s *Server) SetAdmins(admins []keys.ID) {
	s.admins = admins
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
	// e.POST("/task/expired", s.taskExpired)
	e.GET("/task/create/check/:kid", s.createTaskCheck)

	// Cron
	e.POST("/cron/check", s.cronCheck)
	e.POST("/cron/expired", s.cronExpired)

	// Messages
	e.POST("/msgs/:kid/:rid", s.postMessage)
	e.GET("/msgs/:kid/:rid", s.listMessages)

	// Vault
	e.POST("/vault/:kid", s.postVault)
	e.GET("/vault/:kid", s.listVault)
	e.PUT("/vault/:kid", s.putVault)

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

	// Admin
	e.POST("/admin/check/:kid", s.adminCheck)
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
