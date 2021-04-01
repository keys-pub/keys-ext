package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// ErrHTTP ...
type ErrHTTP struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Err     error  `json:"-"`
}

func (e ErrHTTP) Error() string {
	return e.Message
}

type response struct {
	Error *ErrHTTP `json:"error,omitempty"`
}

func newError(code int, err error) ErrHTTP {
	msg := err.Error()
	if code >= 500 {
		// Hide 5xx error messages
		msg = "internal error"
	}
	return ErrHTTP{
		Code:    code,
		Message: msg,
		Err:     err,
	}
}

// ErrResponse is a generate error response.
func (s *Server) ErrResponse(c echo.Context, err error) error {
	s.logger.Errorf("Error: %+v", err)

	switch v := err.(type) {
	case ErrHTTP:
		return JSON(c, v.Code, &response{Error: &v})
	case *ErrHTTP:
		return JSON(c, v.Code, &response{Error: v})
	}

	errh := newError(http.StatusInternalServerError, err)
	return JSON(c, http.StatusInternalServerError, &response{Error: &errh})
}

// ErrBadRequest response.
func (s *Server) ErrBadRequest(c echo.Context, err error) error {
	return s.ErrResponse(c, newError(http.StatusBadRequest, err))
}

// ErrEntityTooLarge response.
func (s *Server) ErrEntityTooLarge(c echo.Context, err error) error {
	return s.ErrResponse(c, newError(http.StatusRequestEntityTooLarge, err))
}

// ErrForbidden response.
func (s *Server) ErrForbidden(c echo.Context, err error) error {
	return s.ErrResponse(c, newError(http.StatusForbidden, err))
}

// ErrTooManyRequests response.
func (s *Server) ErrTooManyRequests(c echo.Context, err error) error {
	return s.ErrResponse(c, newError(http.StatusTooManyRequests, err))
}

// ErrConflict response.
func (s *Server) ErrConflict(c echo.Context, err error) error {
	return s.ErrResponse(c, newError(http.StatusConflict, err))
}

// ErrNotFound response.
func (s *Server) ErrNotFound(c echo.Context, err error) error {
	if err == nil {
		err = errors.Errorf("resource not found")
	}
	return s.ErrResponse(c, newError(http.StatusNotFound, err))
}

// ErrUnauthorized response.
// Use ErrForbidden instead.
// func ErrUnauthorized(c echo.Context, err error) error {
// 	return s.ErrResponse(c, http.StatusUnauthorized, err)
// }

// ErrorHandler returns error handler that returns in the format:
// {"error": {"message": "error message", status: 500}}".
func (s *Server) ErrorHandler(err error, c echo.Context) {
	s.logger.Errorf("Error: %+v", err)

	code := http.StatusInternalServerError
	var resp *response

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg := he.Message
		resp = &response{
			Error: &ErrHTTP{
				Message: strings.ToLower(fmt.Sprintf("%s", msg)),
				Code:    code,
			},
		}
	} else {
		resp = &response{
			Error: &ErrHTTP{
				Message: strings.ToLower(http.StatusText(code)),
				Code:    code,
			},
		}
	}

	// Send response
	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead { // Issue #608
			if err := c.NoContent(code); err != nil {
				s.logger.Errorf("Error (no content): %v", err)
			}
		} else {
			if err := JSON(c, code, resp); err != nil {
				s.logger.Errorf("Error (JSON): %v", err)
			}
		}
	}
}
