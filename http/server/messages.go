package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/encoding"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type message struct {
	ID     string `json:"id"`
	Data   []byte `json:"data"`
	Expire string `json:"exp,omitempty"`
}

func (s *Server) postMessage(c echo.Context) error {
	s.logger.Infof("Server %s %s", c.Request().Method, c.Request().URL.String())

	kid, status, err := authorize(c, s.URL, "kid", s.nowFn(), s.mc)
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

	id := encoding.MustEncode(keys.RandBytes(32), encoding.Base62)

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

	path := ds.Path("msgs", id)
	if err := s.fi.Create(ctx, path, mb); err != nil {
		return s.internalError(c, err)
	}

	spath := ds.Path("direct", kid, rid)
	if err := s.fi.ChangesAdd(ctx, spath, [][]byte{[]byte(path)}); err != nil {
		return s.internalError(c, err)
	}

	if kid != rid {
		rpath := ds.Path("direct", rid, kid)
		if err := s.fi.ChangesAdd(ctx, rpath, [][]byte{[]byte(path)}); err != nil {
			// TODO: This could leave in an inconsistent state (only 1 person sees message)
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
	ctx := c.Request().Context()

	kid, status, err := authorize(c, s.URL, "kid", s.nowFn(), s.mc)
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

	path := ds.Path("direct", kid, rid)

	chgs, clientErr, err := s.changes(c, path)
	if err != nil {
		return s.internalError(c, err)
	}
	if clientErr != nil {
		return ErrResponse(c, http.StatusBadRequest, clientErr.Error())
	}

	docs, err := s.docsFromChanges(ctx, chgs.changes)
	if err != nil {
		return s.internalError(c, err)
	}

	messages := make([]*api.Message, 0, len(chgs.changes))
	md := make(map[string]api.Metadata, len(chgs.changes))
	for _, doc := range docs {
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
	fields := ds.NewStringSetSplit(c.QueryParam("include"), ",")
	if fields.Contains("md") {
		resp.Metadata = md
	}
	return JSON(c, http.StatusOK, resp)
}

func (s *Server) docsFromChanges(ctx context.Context, chgs []*ds.Change) ([]*ds.Document, error) {
	paths := make([]string, 0, len(chgs))
	for _, c := range chgs {
		path := string(c.Data)
		if path != "" {
			paths = append(paths, path)
		}
	}
	return s.fi.GetAll(ctx, paths)
}

func (s *Server) msgFromDoc(doc *ds.Document) (*api.Message, error) {
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

func (s *Server) checkMessage(msg *message, doc *ds.Document) (bool, error) {
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
