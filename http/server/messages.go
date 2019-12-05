package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/api"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

const msgChanges = "msg-changes"

func (s *Server) putMessage(c echo.Context) error {
	request := c.Request()
	ctx := request.Context()
	logger.Infof(ctx, "Server PUT message %s", s.urlString(c))

	// Auth
	auth := request.Header.Get("Authorization")
	if auth == "" {
		return ErrUnauthorized(c, errors.Errorf("missing Authorization header"))
	}
	now := s.nowFn()
	authRes, err := CheckAuthorization(request.Context(), request.Method, s.urlString(c), auth, s.mc, now)
	if err != nil {
		return ErrForbidden(c, err)
	}
	kidAuth := authRes.kid
	// End Auth

	kid, err := keys.ParseID(c.Param("kid"))
	if err != nil {
		return ErrBadRequest(c, err)
	}

	if kid != kidAuth {
		return ErrForbidden(c, errors.Errorf("invalid kid"))
	}

	logname := msgChanges + "-" + kid.String()

	id, err := keys.ParseID(c.Param("id"))
	if err != nil {
		return ErrBadRequest(c, err)
	}

	if c.Request().Body == nil {
		return ErrBadRequest(c, errors.Errorf("missing body"))
	}

	bin, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return internalError(c, err)
	}

	msg := api.Message{
		ID:   id,
		Data: bin,
		Path: keys.Path("messages", id),
	}
	logger.Infof(ctx, "Set %s", msg.Path)
	mb, err := json.Marshal(msg)
	if err != nil {
		return internalError(c, err)
	}
	if err := s.fi.Create(ctx, msg.Path, mb); err != nil {
		return internalError(c, err)
	}
	logger.Infof(ctx, "Add change %s %s", logname, msg.Path)
	if err := s.fi.ChangeAdd(ctx, logname, msg.Path); err != nil {
		return internalError(c, err)
	}

	return c.String(http.StatusOK, "{}")
}

func (s *Server) listMessages(c echo.Context) error {
	request := c.Request()
	ctx := request.Context()
	logger.Infof(ctx, "Server GET messages %s", s.urlString(c))

	// Auth
	auth := request.Header.Get("Authorization")
	if auth == "" {
		return ErrUnauthorized(c, errors.Errorf("missing Authorization header"))
	}
	now := s.nowFn()
	authRes, err := CheckAuthorization(request.Context(), request.Method, s.urlString(c), auth, s.mc, now)
	if err != nil {
		return ErrForbidden(c, err)
	}
	kidAuth := authRes.kid
	// End Auth

	kid, err := keys.ParseID(c.Param("kid"))
	if err != nil {
		return ErrBadRequest(c, err)
	}

	if kid != kidAuth {
		return ErrForbidden(c, errors.Errorf("invalid kid"))
	}
	path := msgChanges + "-" + kid.String()

	le, err := s.changes(c, path)
	if err != nil {
		return internalError(c, err)
	}
	if le.badRequest != nil {
		return le.badRequest
	}
	if len(le.docs) == 0 && le.version == 0 {
		return ErrNotFound(c, errors.Errorf("messages not found"))
	}

	messages := make([]*api.Message, 0, len(le.docs))
	md := make(map[string]api.Metadata, len(le.docs))
	for _, doc := range le.docs {
		var msg api.Message
		if err := json.Unmarshal(doc.Data, &msg); err != nil {
			return internalError(c, err)
		}
		messages = append(messages, &msg)
		md[msg.Path] = api.Metadata{
			CreatedAt: doc.CreatedAt,
			UpdatedAt: doc.UpdatedAt,
		}
	}

	resp := api.MessagesResponse{
		Messages: messages,
		KID:      kid,
		Version:  fmt.Sprintf("%d", le.versionNext),
	}
	fields := keys.NewStringSetSplit(c.QueryParam("include"), ",")
	if fields.Contains("md") {
		resp.Metadata = md
	}
	return JSON(c, http.StatusOK, resp)
}
