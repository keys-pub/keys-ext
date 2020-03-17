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

	recipient := c.Param("rid")
	if recipient == "" {
		return ErrBadRequest(c, errors.Errorf("no recipient id"))
	}
	rid, err := keys.ParseID(recipient)
	if err != nil {
		return ErrBadRequest(c, err)
	}

	channel := c.QueryParam("channel")
	if channel == "" {
		channel = "default"
	}
	if len(channel) > 16 {
		// TODO: Test this
		return ErrBadRequest(c, errors.Errorf("channel name too long"))
	}

	if c.Request().Body == nil {
		return ErrBadRequest(c, errors.Errorf("missing body"))
	}

	// TODO: Limit body size

	id := keys.RandIDString()

	bin, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return internalError(c, err)
	}

	msg := api.Message{
		ID:   id,
		Data: bin,
	}

	mb, err := json.Marshal(msg)
	if err != nil {
		return internalError(c, err)
	}

	path := keys.Path("messages", fmt.Sprintf("%s-%s-%s-%s", kid, rid, channel, id))
	logger.Infof(ctx, "Save message %s", path)
	if err := s.fi.Create(ctx, path, mb); err != nil {
		return internalError(c, err)
	}
	rpath := keys.Path("messages", fmt.Sprintf("%s-%s-%s-%s", rid, kid, channel, id))
	if kid != rid {
		logger.Infof(ctx, "Save message (recipient) %s", rpath)
		if err := s.fi.Create(ctx, rpath, mb); err != nil {
			return internalError(c, err)
		}
	}

	changePath := fmt.Sprintf("%s-%s-%s-%s", msgChanges, kid, rid, channel)
	logger.Infof(ctx, "Add change %s %s", changePath, path)
	if err := s.fi.ChangeAdd(ctx, changePath, path); err != nil {
		return internalError(c, err)
	}
	if kid != rid {
		rchangePath := fmt.Sprintf("%s-%s-%s-%s", msgChanges, rid, kid, channel)
		logger.Infof(ctx, "Add change (recipient) %s %s", rchangePath, rpath)
		if err := s.fi.ChangeAdd(ctx, rchangePath, rpath); err != nil {
			return internalError(c, err)
		}
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

	recipient := c.Param("rid")
	if recipient == "" {
		return ErrBadRequest(c, errors.Errorf("no recipient id"))
	}
	rid, err := keys.ParseID(recipient)
	if err != nil {
		return ErrBadRequest(c, err)
	}

	channel := c.QueryParam("channel")
	if channel == "" {
		channel = "default"
	}
	if len(channel) > 16 {
		// TODO: Test this
		return ErrBadRequest(c, errors.Errorf("channel name too long"))
	}

	changePath := fmt.Sprintf("%s-%s-%s-%s", msgChanges, kid, rid, channel)

	chgs, err := s.changes(c, changePath)
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
		Version:  fmt.Sprintf("%d", chgs.versionNext),
	}
	fields := keys.NewStringSetSplit(c.QueryParam("include"), ",")
	if fields.Contains("md") {
		resp.Metadata = md
	}
	return JSON(c, http.StatusOK, resp)
}
