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
func ErrResponse(c echo.Context, status int, msg string) error {
	c.Logger().Infof("Error (%d): %s", status, msg)
	return JSON(c, status, newErrorResponse(msg, status))
}

// ErrBadRequest response.
func ErrBadRequest(c echo.Context, err error) error {
	return ErrResponse(c, http.StatusBadRequest, err.Error())
}

// ErrEntityTooLarge response.
func ErrEntityTooLarge(c echo.Context, err error) error {
	return ErrResponse(c, http.StatusRequestEntityTooLarge, err.Error())
}

// ErrForbidden response.
func ErrForbidden(c echo.Context, err error) error {
	return ErrResponse(c, http.StatusForbidden, err.Error())
}

// ErrConflict response.
func ErrConflict(c echo.Context, err error) error {
	return ErrResponse(c, http.StatusConflict, err.Error())
}

// ErrNotFound response.
func ErrNotFound(c echo.Context, err error) error {
	if err == nil {
		err = errors.Errorf("resource not found")
	}
	return ErrResponse(c, http.StatusNotFound, err.Error())
}

// ErrUnauthorized response.
func ErrUnauthorized(c echo.Context, err error) error {
	return ErrResponse(c, http.StatusUnauthorized, err.Error())
}

func internalError(c echo.Context, err error) error {
	return ErrResponse(c, http.StatusInternalServerError, err.Error())
}

// ErrorHandler returns error handler that returns in the format:
// {"error": {"message": "error message", status: 500}}".
func ErrorHandler(err error, c echo.Context) {
	c.Logger().Infof("Error: %v", err)

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
