package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/dstore/events"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keys/request"
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

	// secretKey for encrypting between internal services.
	secretKey *[32]byte
}

// Fire defines interface for remote store (like Firestore).
type Fire interface {
	dstore.Documents
	events.Events
}

// New creates a Server.
func New(fi Fire, rds Redis, req request.Requestor, clock tsutil.Clock, logger Logger) *Server {
	sigchains := keys.NewSigchains(fi)
	usrs := users.New(fi, sigchains, users.Requestor(req), users.Clock(clock))
	return &Server{
		fi:        fi,
		rds:       rds,
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

// SetSecretKeyFromHex as (32 byte hex string) for encrypting between internal services.
func (s *Server) SetSecretKeyFromHex(secretKey string) error {
	if secretKey == "" {
		return errors.Errorf("empty secret key")
	}
	sk, err := encoding.Decode(secretKey, encoding.Hex)
	if err != nil {
		return err
	}
	s.secretKey = keys.Bytes32(sk)
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
	e.PUT("/channel/:cid", s.putChannel)                  // Create a channel
	e.GET("/channel/:cid", s.getChannel)                  // Get a channel
	e.GET("/channel/:cid/users", s.getChannelUsers)       // List channel users
	e.POST("/channel/:cid/msgs", s.postMessage)           // Send message
	e.GET("/channel/:cid/msgs", s.listMessages)           // List messages
	e.POST("/channel/:cid/invites", s.postChannelInvites) // Invite to channel
	e.GET("/channel/:cid/invites", s.getChannelInvites)   // List channel invites

	// User (channels)
	e.GET("/user/:kid/channels", s.listUserChannels)              // List channels for user
	e.GET("/user/:kid/invites", s.getUserChannelInvites)          // List invites
	e.GET("/user/:kid/invite/:cid", s.getUserChannelInvite)       // Get invite
	e.DELETE("/user/:kid/invite/:cid", s.deleteUserChannelInvite) // Delete invite
	e.PUT("/user/:kid/channel/:cid", s.putUserChannel)            // Join
	e.DELETE("/user/:kid/channel/:cid", s.deleteUserChannel)      // Leave

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
