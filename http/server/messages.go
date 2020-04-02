package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/keys-pub/keys"
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

	channel := c.Param("channel")
	if channel == "" {
		return ErrBadRequest(c, errors.Errorf("no channel"))
	}
	if len(channel) > 16 {
		return ErrBadRequest(c, errors.Errorf("channel name too long"))
	}

	id := c.Param("id")
	if id == "" {
		return ErrBadRequest(c, errors.Errorf("no id"))
	}
	if len(id) > 64 {
		return ErrBadRequest(c, errors.Errorf("id too long"))
	}

	expire := time.Hour * 24
	if c.QueryParam("expire") != "" {
		e, err := time.ParseDuration(c.QueryParam("expire"))
		if err != nil {
			return ErrBadRequest(c, err)
		}
		expire = e
	}
	if len(expire.String()) > 64 {
		return ErrBadRequest(c, errors.Errorf("invalid expire"))
	}
	if expire > time.Hour*24 {
		return ErrBadRequest(c, errors.Errorf("max expire is 24h"))
	}

	if c.Request().Body == nil {
		return ErrBadRequest(c, errors.Errorf("missing body"))
	}

	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return s.internalError(c, err)
	}

	if len(b) > 16*1024 {
		// TODO: Check length before reading data
		return ErrBadRequest(c, errors.Errorf("message too large (greater than 16KiB)"))
	}

	return s.saveMessage(c, kid, rid, channel, id, b, expire)
}

type message struct {
	ID     string `json:"id"`
	Data   []byte `json:"data"`
	Expire string `json:"exp,omitempty"`
}

func (s *Server) saveMessage(c echo.Context, kid keys.ID, rid keys.ID, channel string, id string, b []byte, expire time.Duration) error {
	ctx := c.Request().Context()

	exp := expire.String()
	if expire == time.Duration(0) {
		exp = ""
	}

	msg := message{
		ID:     id,
		Data:   b,
		Expire: exp,
	}
	mb, err := json.Marshal(msg)
	if err != nil {
		return s.internalError(c, err)
	}

	path := keys.Path("messages", keys.Rand3262())
	if err := s.fi.Set(ctx, path, mb); err != nil {
		return s.internalError(c, err)
	}

	currpath := keys.Path("messages", fmt.Sprintf("%s-%s-%s-%s", kid, rid, channel, id))
	s.logger.Infof("Save message %s", path)
	if err := s.fi.Set(ctx, currpath, []byte(path)); err != nil {
		return s.internalError(c, err)
	}
	rpath := keys.Path("messages", fmt.Sprintf("%s-%s-%s-%s", rid, kid, channel, id))
	if kid != rid {
		s.logger.Infof("Save message (recipient) %s", rpath)
		if err := s.fi.Set(ctx, rpath, []byte(path)); err != nil {
			return s.internalError(c, err)
		}
	}

	changePath := fmt.Sprintf("%s-%s-%s-%s", msgChanges, kid, rid, channel)
	s.logger.Infof("Add change %s %s", changePath, path)
	if err := s.fi.ChangeAdd(ctx, changePath, path); err != nil {
		return s.internalError(c, err)
	}
	if kid != rid {
		rchangePath := fmt.Sprintf("%s-%s-%s-%s", msgChanges, rid, kid, channel)
		s.logger.Infof("Add change (recipient) %s %s", rchangePath, path)
		if err := s.fi.ChangeAdd(ctx, rchangePath, path); err != nil {
			return s.internalError(c, err)
		}
	}

	var resp struct{}
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

	channel := c.Param("channel")
	if channel == "" {
		return ErrBadRequest(c, errors.Errorf("no channel"))
	}
	if len(channel) > 16 {
		return ErrBadRequest(c, errors.Errorf("channel name too long"))
	}

	changePath := fmt.Sprintf("%s-%s-%s-%s", msgChanges, kid, rid, channel)

	chgs, err := s.changes(c, changePath)
	if err != nil {
		return s.internalError(c, err)
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
		msg, err := s.msgFromDoc(doc)
		if err != nil {
			return s.internalError(c, err)
		}
		if msg == nil {
			continue
		}

		messages = append(messages, msg)
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

func (s *Server) msgFromDoc(doc *keys.Document) (*api.Message, error) {
	if doc == nil {
		return nil, nil
	}
	var msg message
	if err := json.Unmarshal(doc.Data, &msg); err != nil {
		return nil, err
	}
	ok, err := s.checkMessage(&msg, doc)
	if err != nil {
		return nil, err
	}
	if !ok {
		s.logger.Debugf("Message pruned: %s", doc.Path)
		return nil, nil
	}
	return &api.Message{
		ID:   msg.ID,
		Data: msg.Data,
	}, nil
}

func (s *Server) checkMessage(msg *message, doc *keys.Document) (bool, error) {
	if msg.Expire != "" {
		expiry, err := time.ParseDuration(msg.Expire)
		if err != nil {
			return false, errors.Wrapf(err, "invalid expire")
		}
		now := s.nowFn()
		s.logger.Debugf("Now: %s, Created: %s, Expiry: %s, Sub: %s", now, doc.CreatedAt, expiry, now.Sub(doc.CreatedAt))
		if now.Sub(doc.CreatedAt) > expiry {
			return false, nil
		}
	}
	return true, nil
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

	channel := c.Param("channel")
	if channel == "" {
		return ErrBadRequest(c, errors.Errorf("no channel"))
	}
	if len(channel) > 16 {
		return ErrBadRequest(c, errors.Errorf("channel name too long"))
	}

	id := c.Param("id")
	if id == "" {
		return ErrBadRequest(c, errors.Errorf("no id"))
	}
	if len(id) > 64 {
		return ErrBadRequest(c, errors.Errorf("id too long"))
	}

	path := keys.Path("messages", fmt.Sprintf("%s-%s-%s-%s", kid, rid, channel, id))
	ok, err := s.fi.Delete(ctx, path)
	if err != nil {
		return s.internalError(c, err)
	}
	rpath := keys.Path("messages", fmt.Sprintf("%s-%s-%s-%s", rid, kid, channel, id))
	rok, err := s.fi.Delete(ctx, rpath)
	if err != nil {
		return s.internalError(c, err)
	}
	if !ok && !rok {
		return ErrNotFound(c, errors.Errorf("message not found"))
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

	channel := c.Param("channel")
	if channel == "" {
		return ErrBadRequest(c, errors.Errorf("no channel"))
	}
	if len(channel) > 16 {
		return ErrBadRequest(c, errors.Errorf("channel name too long"))
	}

	id := c.Param("id")
	if id == "" {
		return ErrBadRequest(c, errors.Errorf("no id"))
	}
	if len(id) > 64 {
		return ErrBadRequest(c, errors.Errorf("id too long"))
	}

	path := keys.Path("messages", fmt.Sprintf("%s-%s-%s-%s", kid, rid, channel, id))
	docRef, err := s.fi.Get(ctx, path)
	if err != nil {
		return s.internalError(c, err)
	}
	if docRef == nil {
		return ErrNotFound(c, errors.Errorf("message not found"))
	}
	doc, err := s.fi.Get(ctx, string(docRef.Data))
	if err != nil {
		return s.internalError(c, err)
	}

	msg, err := s.msgFromDoc(doc)
	if err != nil {
		return s.internalError(c, err)
	}
	if msg == nil {
		return ErrNotFound(c, errors.Errorf("message not found"))
	}

	return c.Blob(http.StatusOK, echo.MIMEOctetStream, msg.Data)
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
