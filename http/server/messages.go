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

type message struct {
	ID     string `json:"id"`
	Data   []byte `json:"data"`
	Expire string `json:"exp,omitempty"`
}

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

	exp := expire.String()
	if expire == time.Duration(0) {
		exp = ""
	}

	id := keys.Rand3262()

	msg := message{
		ID:     id,
		Data:   b,
		Expire: exp,
	}
	mb, err := json.Marshal(msg)
	if err != nil {
		return s.internalError(c, err)
	}

	ctx := c.Request().Context()

	path := keys.Path("msgs", id)
	if err := s.fi.Create(ctx, path, mb); err != nil {
		return s.internalError(c, err)
	}

	spath := fmt.Sprintf("msgs-%s-%s", kid, rid)
	if err := s.fi.ChangeAdd(ctx, spath, id, path); err != nil {
		return s.internalError(c, err)
	}

	if kid != rid {
		rpath := fmt.Sprintf("msgs-%s-%s", rid, kid)
		if err := s.fi.ChangeAdd(ctx, rpath, id, path); err != nil {
			return s.internalError(c, err)
		}
	}

	resp := api.CreateMessageResponse{
		ID: id,
	}
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

	path := fmt.Sprintf("msgs-%s-%s", kid, rid)

	chgs, err := s.changes(c, path)
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

// func truthy(s string) bool {
// 	s = strings.TrimSpace(s)
// 	switch s {
// 	case "", "0", "f", "false", "n", "no":
// 		return false
// 	default:
// 		return true
// 	}
// }
