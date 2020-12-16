package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type errResponse struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type response struct {
	Error *errResponse `json:"error,omitempty"`
}

func newErrorResponse(msg string, code int) *response {
	return &response{
		Error: &errResponse{
			Message: msg,
			Code:    code,
		},
	}
}

// ErrResponse is a generate error response.
func ErrResponse(c echo.Context, status int, err error) error {
	return JSON(c, status, newErrorResponse(err.Error(), status))
}

// ErrInternalServer response.
func ErrInternalServer(c echo.Context, err error) error {
	return ErrResponse(c, http.StatusInternalServerError, err)
}

// ErrBadRequest response.
func ErrBadRequest(c echo.Context, err error) error {
	return ErrResponse(c, http.StatusBadRequest, err)
}

// ErrEntityTooLarge response.
func ErrEntityTooLarge(c echo.Context, err error) error {
	return ErrResponse(c, http.StatusRequestEntityTooLarge, err)
}

// ErrForbidden response.
func ErrForbidden(c echo.Context, err error) error {
	// We hide the source of the error to not expose any metadata.
	return ErrResponse(c, http.StatusForbidden, errors.Errorf("auth failed"))
}

// ErrConflict response.
func ErrConflict(c echo.Context, err error) error {
	return ErrResponse(c, http.StatusConflict, err)
}

// ErrNotFound response.
func ErrNotFound(c echo.Context, err error) error {
	if err == nil {
		err = errors.Errorf("resource not found")
	}
	return ErrResponse(c, http.StatusNotFound, err)
}

// ErrUnauthorized response.
// Use ErrForbidden instead.
// func ErrUnauthorized(c echo.Context, err error) error {
// 	return ErrResponse(c, http.StatusUnauthorized, err)
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
			Error: &errResponse{
				Message: strings.ToLower(fmt.Sprintf("%s", msg)),
				Code:    code,
			},
		}
	} else {
		resp = &response{
			Error: &errResponse{
				Message: strings.ToLower(http.StatusText(code)),
				Code:    code,
			},
		}
	}

	// Send response
	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead { // Issue #608
			if err := c.NoContent(code); err != nil {
				c.Logger().Errorf("Error (no content): %v", err)
			}
		} else {
			if err := JSON(c, code, resp); err != nil {
				c.Logger().Errorf("Error (JSON): %v", err)
			}
		}
	}
}
