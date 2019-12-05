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
	logger.Warningf(c.Request().Context(), "Error (%d): %s", status, msg)
	return JSON(c, status, newErrorResponse(msg, status))
}

// ErrBadRequest response.
func ErrBadRequest(c echo.Context, err error) error {
	logger.Warningf(c.Request().Context(), "Bad request (400): %v", err)
	return JSON(c, http.StatusBadRequest, newErrorResponse(err.Error(), http.StatusBadRequest))
}

// ErrEntityTooLarge response.
func ErrEntityTooLarge(c echo.Context, err error) error {
	logger.Warningf(c.Request().Context(), "Entity too large (413): %v", err)
	return JSON(c, http.StatusRequestEntityTooLarge, newErrorResponse(err.Error(), http.StatusRequestEntityTooLarge))
}

// ErrForbidden response.
func ErrForbidden(c echo.Context, err error) error {
	logger.Warningf(c.Request().Context(), "Forbidden (403): %v", err)
	return JSON(c, http.StatusForbidden, newErrorResponse(err.Error(), http.StatusForbidden))
}

// ErrConflict response.
func ErrConflict(c echo.Context, err error) error {
	logger.Warningf(c.Request().Context(), "Conflict (409): %v", err)
	return JSON(c, http.StatusConflict, newErrorResponse(err.Error(), http.StatusConflict))
}

// ErrNotFound response.
func ErrNotFound(c echo.Context, err error) error {
	if err == nil {
		err = errors.Errorf("resource not found")
	}
	logger.Warningf(c.Request().Context(), "Not found: %s", err)
	return JSON(c, http.StatusNotFound, newErrorResponse(err.Error(), http.StatusNotFound))
}

// ErrUnauthorized response.
func ErrUnauthorized(c echo.Context, err error) error {
	logger.Warningf(c.Request().Context(), "Bad auth: %v", err)
	return JSON(c, http.StatusUnauthorized, newErrorResponse(err.Error(), http.StatusUnauthorized))
}

func internalError(c echo.Context, err error) error {
	logger.Errorf(c.Request().Context(), "Error (internal): %v", err)
	return JSON(c, http.StatusInternalServerError, newErrorResponse(err.Error(), http.StatusInternalServerError))
}

// ErrorHandler returns error handler that returns in the format:
// {"error": {"message": "error message", status: 500}}".
func ErrorHandler(err error, c echo.Context) {
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
				c.Logger().Error(err)
			}
		} else {
			if err := JSON(c, code, resp); err != nil {
				c.Logger().Error(err)
			}
		}
	}
}
