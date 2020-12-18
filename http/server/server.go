package server

import (
	"encoding/json"
	"io/ioutil"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/dstore/events"
	"github.com/keys-pub/keys/http"
	"github.com/keys-pub/keys/tsutil"
	"github.com/keys-pub/keys/users"
	"github.com/labstack/echo/v4"

	"github.com/pkg/errors"
)

// TODO: Support If-Modified-Since

// TODO: Turn off logging

// Server ...
type Server struct {
	fi     Fire
	rds    Redis
	clock  tsutil.Clock
	logger Logger
	client http.Client

	// URL (base) of form http(s)://host:port with no trailing slash to help
	// authorization checks in testing where the host is ambiguous.
	URL string

	users     *users.Users
	sigchains *keys.Sigchains
	tasks     Tasks

	// internalAuth token for authorizing internal services.
	internalAuth string

	// admins are key ids that can do admin actions on the server.
	admins []keys.ID
}

// Fire defines interface for remote store (like Firestore).
type Fire interface {
	dstore.Documents
	events.Events
}

// New creates a Server.
func New(fi Fire, rds Redis, client http.Client, clock tsutil.Clock, logger Logger) *Server {
	sigchains := keys.NewSigchains(fi)

	usrs := users.New(fi, sigchains, users.Client(client), users.Clock(clock))
	return &Server{
		fi:        fi,
		rds:       rds,
		client:    client,
		clock:     tsutil.NewClock(),
		tasks:     newUnsetTasks(),
		sigchains: sigchains,
		users:     usrs,
		logger:    logger,
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
	e.HTTPErrorHandler = s.ErrorHandler
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

	// Cron
	e.POST("/cron/check", s.cronCheck)

	// Vault
	e.POST("/vault/:kid", s.postVault)
	e.GET("/vault/:kid", s.listVault)
	e.DELETE("/vault/:kid", s.deleteVault)
	e.HEAD("/vault/:kid", s.headVault)

	// Disco
	e.PUT("/disco/:kid/:rid/:type", s.putDisco)
	e.GET("/disco/:kid/:rid/:type", s.getDisco)
	e.DELETE("/disco/:kid/:rid", s.deleteDisco)

	// Invite Code
	e.POST("/invite/code/:kid/:rid", s.postInviteCode)
	e.GET("/invite/code/:code", s.getInviteCode)

	// Share
	e.GET("/share/:kid", s.getShare)
	e.PUT("/share/:kid", s.putShare)

	// Sigchain (aliases)
	e.GET("/:kid", s.getSigchainAliased)
	e.GET("/:kid/:seq", s.getSigchainStatementAliased)
	e.PUT("/:kid/:seq", s.putSigchainStatementAliased)

	//
	// Experimental
	//

	// Channel
	e.PUT("/channel/:cid", s.putChannel)        // Create a channel
	e.GET("/channel/:cid", s.getChannel)        // Get a channel
	e.POST("/channel/:cid/msgs", s.postMessage) // Send message
	e.GET("/channel/:cid/msgs", s.getMessages)  // List messages

	// Direct
	e.POST("/dm/:to/:from", s.postDirect) // Send direct message to :kid from :sid
	e.GET("/dm/:kid", s.getDirects)       // List direct messages

	// Follow
	e.POST("/follow/:kid/:user", s.postFollow)     // Follow
	e.GET("/follows/:kid", s.getFollows)           // List follows
	e.DELETE("/follow/:kid/:user", s.deleteFollow) // Unfollow

	// Twitter
	e.POST("/twitter/check/:name/:id", s.checkTwitter)

	// Admin
	e.POST("/admin/check/:kid", s.adminCheck)
}

// SetClock sets clock.
func (s *Server) SetClock(clock tsutil.Clock) {
	s.clock = clock
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
		return s.ErrForbidden(c, errors.Errorf("no auth token set on server"))
	}
	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		return s.ErrForbidden(c, errors.Errorf("no auth token specified"))
	}
	if auth != s.internalAuth {
		return s.ErrForbidden(c, errors.Errorf("invalid auth token"))
	}
	return nil
}

func readBody(c echo.Context, required bool, maxLength int) ([]byte, int, error) {
	br := c.Request().Body
	if br == nil {
		if !required {
			return []byte{}, 0, nil
		}
		return nil, http.StatusBadRequest, errors.Errorf("missing body")
	}
	b, err := ioutil.ReadAll(br)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if len(b) > maxLength {
		// TODO: Check length before reading data
		return nil, http.StatusRequestEntityTooLarge, errors.Errorf("request too large")
	}
	if len(b) == 0 && required {
		return nil, http.StatusBadRequest, errors.Errorf("no body data")
	}
	return b, 0, nil
}
