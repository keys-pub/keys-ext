package server

import (
	"encoding/json"
	"io/ioutil"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/dstore/events"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keys/http"
	"github.com/keys-pub/keys/tsutil"
	"github.com/keys-pub/keys/users"
	"github.com/labstack/echo/v4"
	"github.com/vmihailenco/msgpack/v4"

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

	// internalKey for encrypting between internal services.
	internalKey *[32]byte

	// tokenKey for JWT vault tokens
	tokenKey []byte

	emailer Emailer
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
		clock:     clock,
		tasks:     newUnsetTasks(),
		sigchains: sigchains,
		users:     usrs,
		logger:    logger,
	}
}

// Emailer sends emails.
type Emailer interface {
	SendVerificationEmail(email string, code string) error
}

// SetInternalAuth for authorizing internal requests, like tasks.
func (s *Server) SetInternalAuth(internalAuth string) {
	s.internalAuth = internalAuth
}

// SetTokenKey for setting token key.
func (s *Server) SetTokenKey(tokenKey string) error {
	if tokenKey == "" {
		return errors.Errorf("empty token key")
	}
	k, err := encoding.Decode(tokenKey, encoding.Hex)
	if err != nil {
		return err
	}
	s.tokenKey = k
	return nil
}

// SetEmailer sets emailer.
func (s *Server) SetEmailer(emailer Emailer) {
	s.emailer = emailer
}

// SetInternalKey for encrypting between internal services.
func (s *Server) SetInternalKey(internalKey string) error {
	if internalKey == "" {
		return errors.Errorf("empty internal key")
	}
	sk, err := encoding.Decode(internalKey, encoding.Hex)
	if err != nil {
		return err
	}
	s.internalKey = keys.Bytes32(sk)
	return nil
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
	e.GET("/user/:user", s.getUser)

	// Tasks
	e.POST("/task/check/:kid", s.taskCheck)

	// Cron
	e.POST("/cron/check", s.cronCheck)

	// Vault
	e.POST("/vault/:vid", s.postVault)
	e.GET("/vault/:vid", s.listVault)
	e.DELETE("/vault/:vid", s.deleteVault)
	e.HEAD("/vault/:vid", s.headVault)

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

	// Batch
	e.POST("/batch", s.postBatch) // Batch

	// Twitter
	e.GET("/twitter/:kid/:name/:id", s.checkTwitter)

	// Admin
	e.POST("/admin/check/:kid", s.adminCheck)

	// Sigchain (aliases)
	e.GET("/:kid", s.getSigchainAliased)
	e.GET("/:kid/:seq", s.getSigchainStatementAliased)
	e.PUT("/:kid/:seq", s.putSigchainStatementAliased)
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

// Msgpack response.
func Msgpack(c echo.Context, status int, i interface{}) error {
	var b []byte
	switch v := i.(type) {
	case []byte:
		b = v
	default:
		mb, err := msgpack.Marshal(i)
		if err != nil {
			panic(err)
		}
		b = mb
	}
	return c.Blob(status, echo.MIMEApplicationMsgpack, b)
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

func readBody(c echo.Context, required bool, maxLength int) ([]byte, error) {
	br := c.Request().Body
	if br == nil {
		if !required {
			return []byte{}, nil
		}
		return nil, newError(http.StatusBadRequest, errors.Errorf("missing body"))
	}
	b, err := ioutil.ReadAll(br)
	if err != nil {
		return nil, newError(http.StatusInternalServerError, err)
	}
	if len(b) > maxLength {
		// TODO: Check length before reading data
		return nil, newError(http.StatusRequestEntityTooLarge, errors.Errorf("request too large"))
	}
	if len(b) == 0 && required {
		return nil, newError(http.StatusBadRequest, errors.Errorf("no body data"))
	}
	return b, nil
}
