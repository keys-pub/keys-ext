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

// TODO: Message expiry

const msgChanges = "msg-changes"

func (s *Server) postMessage(c echo.Context) error {
	request := c.Request()
	ctx := request.Context()
	logger.Infof(ctx, "Server POST message %s", s.urlString(c))

	kid, status, err := s.authorize(c)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	logname := msgChanges + "-" + kid.String()

	id := keys.RandIDString()

	if c.Request().Body == nil {
		return ErrBadRequest(c, errors.Errorf("missing body"))
	}

	bin, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return internalError(c, err)
	}

	path := keys.Path("messages", fmt.Sprintf("%s-%s", kid, id))

	msg := api.Message{
		ID:   id,
		Data: bin,
	}
	logger.Infof(ctx, "Save message %s", path)
	mb, err := json.Marshal(msg)
	if err != nil {
		return internalError(c, err)
	}
	if err := s.fi.Create(ctx, path, mb); err != nil {
		return internalError(c, err)
	}
	logger.Infof(ctx, "Add change %s %s", logname, path)
	if err := s.fi.ChangeAdd(ctx, logname, path); err != nil {
		return internalError(c, err)
	}

	resp := api.MessageResponse{
		ID: id,
	}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) listMessages(c echo.Context) error {
	request := c.Request()
	ctx := request.Context()
	logger.Infof(ctx, "Server GET messages %s", s.urlString(c))

	kid, status, err := s.authorize(c)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	path := msgChanges + "-" + kid.String()

	chgs, err := s.changes(c, path)
	if err != nil {
		return internalError(c, err)
	}
	if chgs.errBadRequest != nil {
		return ErrResponse(c, http.StatusBadRequest, chgs.errBadRequest.Error())
	}
	if len(chgs.docs) == 0 && chgs.version == 0 {
		return ErrNotFound(c, errors.Errorf("messages not found"))
	}

	messages := make([]*api.Message, 0, len(chgs.docs))
	md := make(map[string]api.Metadata, len(chgs.docs))
	for _, doc := range chgs.docs {
		var msg api.Message
		if err := json.Unmarshal(doc.Data, &msg); err != nil {
			return internalError(c, err)
		}
		messages = append(messages, &msg)
		md[msg.ID] = api.Metadata{
			CreatedAt: doc.CreatedAt,
			UpdatedAt: doc.UpdatedAt,
		}
	}

	resp := api.MessagesResponse{
		Messages: messages,
		KID:      kid,
		Version:  fmt.Sprintf("%d", chgs.versionNext),
	}
	fields := keys.NewStringSetSplit(c.QueryParam("include"), ",")
	if fields.Contains("md") {
		resp.Metadata = md
	}
	return JSON(c, http.StatusOK, resp)
}
