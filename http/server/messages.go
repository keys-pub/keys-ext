package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keysd/http/api"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// TODO: Message expiry

const msgChanges = "msg-changes"

func (s *Server) putMessage(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	kid, status, err := authorize(c, s.URL, s.nowFn(), s.mc)
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

	id := c.Param("id")
	if id == "" {
		return ErrBadRequest(c, errors.Errorf("no id"))
	}
	if len(id) > 64 {
		return ErrBadRequest(c, errors.Errorf("id too long"))
	}

	expire := time.Duration(0)
	pexpire := c.QueryParam("expire")
	if pexpire != "" {
		expire, err = time.ParseDuration(pexpire)
		if err != nil {
			return ErrBadRequest(c, err)
		}
	}
	if expire > time.Hour {
		return ErrBadRequest(c, errors.Errorf("max expire is 1h"))
	}

	// channel := c.QueryParam("channel")
	// if len(channel) > 16 {
	// 	return ErrBadRequest(c, errors.Errorf("channel name too long"))
	// }

	if c.Request().Body == nil {
		return ErrBadRequest(c, errors.Errorf("missing body"))
	}

	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return internalError(c, err)
	}

	if len(b) > 512*1024 {
		// TODO: Check length before reading data
		return ErrBadRequest(c, errors.Errorf("message too large (greater than 512KiB)"))
	}

	if expire > 0 {
		return s.expiringMessage(c, kid, rid, expire, id, b)
	} else {
		return s.saveMessage(c, kid, rid, id, b)
	}
}

func (s *Server) saveMessage(c echo.Context, kid keys.ID, rid keys.ID, id string, b []byte) error {
	ctx := c.Request().Context()

	msg := api.Message{
		ID:   id,
		Data: b,
	}

	mb, err := json.Marshal(msg)
	if err != nil {
		return internalError(c, err)
	}

	path := keys.Path("messages", fmt.Sprintf("%s-%s-%s", kid, rid, id))
	s.logger.Infof("Save message %s", path)
	if err := s.fi.Create(ctx, path, mb); err != nil {
		return internalError(c, err)
	}
	rpath := keys.Path("messages", fmt.Sprintf("%s-%s-%s", rid, kid, id))
	if kid != rid {
		s.logger.Infof("Save message (recipient) %s", rpath)
		if err := s.fi.Create(ctx, rpath, mb); err != nil {
			return internalError(c, err)
		}
	}

	changePath := fmt.Sprintf("%s-%s-%s", msgChanges, kid, rid)
	s.logger.Infof("Add change %s %s", changePath, path)
	if err := s.fi.ChangeAdd(ctx, changePath, path); err != nil {
		return internalError(c, err)
	}
	if kid != rid {
		rchangePath := fmt.Sprintf("%s-%s-%s", msgChanges, rid, kid)
		s.logger.Infof("Add change (recipient) %s %s", rchangePath, rpath)
		if err := s.fi.ChangeAdd(ctx, rchangePath, rpath); err != nil {
			return internalError(c, err)
		}
	}

	resp := api.MessageResponse{}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) listMessages(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	kid, status, err := authorize(c, s.URL, s.nowFn(), s.mc)
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

	// channel := c.QueryParam("channel")
	// if channel == "" {
	// 	channel = "default"
	// }
	// if len(channel) > 16 {
	// 	return ErrBadRequest(c, errors.Errorf("channel name too long"))
	// }

	changePath := fmt.Sprintf("%s-%s-%s", msgChanges, kid, rid)

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

func (s *Server) expiringMessage(c echo.Context, kid keys.ID, rid keys.ID, expire time.Duration, id string, b []byte) error {
	ctx := c.Request().Context()

	enc, err := encoding.Encode(b, encoding.Base64)
	if err != nil {
		return internalError(c, err)
	}

	key := fmt.Sprintf("msg-%s-%s-%s", kid, rid, id)
	if err := s.mc.Set(ctx, key, enc); err != nil {
		return internalError(c, err)
	}
	// TODO: Configurable expiry?
	if err := s.mc.Expire(ctx, key, expire); err != nil {
		return internalError(c, err)
	}

	resp := api.MessageResponse{}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) deleteMessage(c echo.Context) error {
	ctx := c.Request().Context()
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	kid, status, err := authorize(c, s.URL, s.nowFn(), s.mc)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	recipient := c.Param("rid")
	if recipient == "" {
		return ErrBadRequest(c, errors.Errorf("no recipient id specified"))
	}
	rid, err := keys.ParseID(recipient)
	if err != nil {
		return ErrBadRequest(c, err)
	}

	id := c.Param("id")
	if id == "" {
		return ErrBadRequest(c, errors.Errorf("no id"))
	}
	if len(id) > 64 {
		return ErrBadRequest(c, errors.Errorf("id too long"))
	}

	// TODO: This only deletes expiring message

	key := fmt.Sprintf("msg-%s-%s-%s", kid, rid, id)
	if err := s.mc.Delete(ctx, key); err != nil {
		return internalError(c, err)
	}

	var resp struct{}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) getMessage(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())
	ctx := c.Request().Context()

	kid, status, err := authorize(c, s.URL, s.nowFn(), s.mc)
	if err != nil {
		return ErrResponse(c, status, err.Error())
	}

	recipient := c.Param("rid")
	if recipient == "" {
		return ErrBadRequest(c, errors.Errorf("no recipient id specified"))
	}
	rid, err := keys.ParseID(recipient)
	if err != nil {
		return ErrBadRequest(c, err)
	}
	id := c.Param("id")
	if id == "" {
		return ErrBadRequest(c, errors.Errorf("no id"))
	}
	if len(id) > 64 {
		return ErrBadRequest(c, errors.Errorf("id too long"))
	}

	// TODO: Only gets expiring message

	key := fmt.Sprintf("msg-%s-%s-%s", rid, kid, id)
	out, err := s.mc.Get(ctx, key)
	if err != nil {
		return internalError(c, err)
	}
	if out == "" {
		return ErrNotFound(c, nil)
	}
	if err := s.mc.Delete(ctx, key); err != nil {
		return internalError(c, err)
	}

	b, err := encoding.Decode(out, encoding.Base64)
	if err != nil {
		return internalError(c, err)
	}
	return c.Blob(http.StatusOK, echo.MIMEOctetStream, b)
}

// func truthy(s string) bool {
// 	s = strings.TrimSpace(s)
// 	switch s {
// 	case "", "0", "f", "false", "n", "no":
// 		return false
// 	default:
// 		return true
// 	}
// }
