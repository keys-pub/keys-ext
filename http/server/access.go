package server

import (
	"fmt"
	"net/http"
)

// AccessResource is resource for access control.
type AccessResource string

const (
	// SigchainResource for sigchain.
	SigchainResource AccessResource = "sigchain"
)

func (r AccessResource) String() string {
	return string(r)
}

// AccessAction is action for access control.
type AccessAction string

const (
	// Put action.
	Put AccessAction = "put"
	// Post action.
	Post AccessAction = "post"
)

// Access returns whether to allow or deny.
type Access struct {
	Allow   bool
	Message string
	// StatusCode (optional) for custom HTTP status (if denied)
	StatusCode int
}

// AccessContext is context for request.
type AccessContext interface {
	RealIP() string
	Request() *http.Request
}

// AccessAllow allow access.
func AccessAllow() Access {
	return Access{Allow: true}
}

// AccessDeny deny access (bad request).
func AccessDeny(msg string) Access {
	return Access{Allow: false, Message: msg, StatusCode: http.StatusBadRequest}
}

// AccessDenyTooManyRequests deny access (too many requests).
func AccessDenyTooManyRequests(msg string) Access {
	if msg == "" {
		msg = "too many requests"
	}
	a := AccessDeny(msg)
	a.StatusCode = http.StatusTooManyRequests
	return a
}

// AccessDenyErrored deny access (an error occurred trying to determine access).
func AccessDenyErrored(err error) Access {
	a := AccessDeny(fmt.Sprintf("error: %s", err))
	a.StatusCode = http.StatusInternalServerError
	return a
}

// AccessFn returns access to resource.
// If error message begins with an integer, it will use that as the http status code.
// For example, "429: too many requests".
type AccessFn func(c AccessContext, resource AccessResource, action AccessAction) Access

// SetAccessFn sets access control.
func (s *Server) SetAccessFn(fn AccessFn) {
	s.accessFn = fn
}
